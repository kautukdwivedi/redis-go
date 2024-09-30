package rdb

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/internal/storage"
)

const (
	rdbHeader          = "REDIS"
	metadataHeader     = 0xFA
	dbHeader           = 0xFE
	keyExpiryHeaderMs  = 0xFC
	keyExpiryHeaderSec = 0xFD
	eofHeader          = 0xFF
)

type RDBFile struct {
	Dir        string
	DBFilename string
	Version    string
	Dbs        []*RDBDatabase
}

type RDBDatabase struct {
	Number              int
	HashTableSize       int
	ExpiryHashTableSize int
	Data                map[string]storage.ExpiringValue
	dataMu              *sync.Mutex
}

func NewRDBFile(dir, dbFilename string) *RDBFile {
	return &RDBFile{
		Dir:        dir,
		DBFilename: dbFilename,
		Dbs:        make([]*RDBDatabase, 0),
	}
}

func (rdbf *RDBFile) Parse() error {
	data, err := rdbf.Read()
	if err != nil {
		return err
	}

	isValid, err := isRDBDataValid(data)
	if err != nil {
		return err
	}
	if !isValid {
		return errors.New("invalid rdb file")
	}

	rdbf.Version = string(data[5:9])

	err = rdbf.parseDbsSection(data)
	if err != nil {
		return err
	}

	return nil
}

func (rdbf *RDBFile) Read() ([]byte, error) {
	data, err := os.ReadFile(fmt.Sprint(rdbf.Dir, "/", rdbf.DBFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to read RDB file: %s", err.Error())
	}
	return data, nil
}

func (rdbf *RDBFile) parseDbsSection(data []byte) error {
	if dataIsMissingHeader(data, dbHeader) {
		return nil
	}

	startIdx := bytes.Index(data, []byte{dbHeader}) + 1

	for {
		dbNumber, bytesConsumed, err := parseSizeEncoding(data[startIdx:])
		if err != nil {
			return fmt.Errorf("parsing error: %s", err.Error())
		}

		startIdx += bytesConsumed + 1

		hashTableSize, bytesConsumed, err := parseSizeEncoding(data[startIdx:])
		if err != nil {
			return fmt.Errorf("parsing error: %s", err.Error())
		}

		startIdx += bytesConsumed

		expiryHashTableSize, bytesConsumed, err := parseSizeEncoding(data[startIdx:])
		if err != nil {
			return fmt.Errorf("parsing error: %s", err.Error())
		}

		startIdx += bytesConsumed

		db := &RDBDatabase{
			Number:              dbNumber,
			HashTableSize:       hashTableSize,
			ExpiryHashTableSize: expiryHashTableSize,
			Data:                make(map[string]storage.ExpiringValue),
			dataMu:              &sync.Mutex{},
		}

		var count int
		var key string

		for {
			str, bytesConsumed, err := stringEncoding(data[startIdx:])
			if err != nil {
				return fmt.Errorf("parsing error: %s", err.Error())
			}

			startIdx += bytesConsumed

			if len(str) > 0 {
				if count%2 == 0 {
					key = str
				} else {
					db.dataMu.Lock()
					db.Data[key] = storage.ExpiringValue{Val: str}
					db.dataMu.Unlock()
				}
				count++
			}
			if startIdx >= len(data) || data[startIdx] == dbHeader || data[startIdx] == eofHeader {
				break
			}
		}

		rdbf.Dbs = append(rdbf.Dbs, db)

		if startIdx >= len(data) || data[startIdx] == dbHeader || data[startIdx] == eofHeader {
			break
		}
	}

	return nil
}

func isRDBDataValid(data []byte) (bool, error) {
	if len(data) < 5 || !bytes.Equal(data[:5], []byte(rdbHeader)) {
		return false, fmt.Errorf("file does not start with %s header", rdbHeader)
	}

	if len(data) < 9 {
		return false, errors.New("version is either missing or invalid")
	}

	if dataIsMissingHeader(data, eofHeader) {
		return false, errors.New("EOF is missing")
	}

	return true, nil
}

func dataIsMissingHeader(data []byte, header byte) bool {
	return !bytes.Contains(data, []byte{header})
}

func parseSizeEncoding(data []byte) (dbNumber, bytesConsumed int, err error) {
	firstByte := data[0]
	firstTwoSignificantBits := (firstByte & 0b11000000)

	switch firstTwoSignificantBits {
	case byte(0b00000000):
		return int(firstByte & 0b00111111), 1, nil
	case byte(0b01000000):
		secondByte := data[1]
		return int((firstByte&0b00111111)<<8 | secondByte), 2, nil
	case byte(0b10000000):
		secondByte := data[1]
		thirdByte := data[2]
		fourthByte := data[3]
		fifthByte := data[4]
		return int(secondByte<<24 | thirdByte<<16 | fourthByte<<8 | fifthByte), 5, nil
	case byte(0b11000000):
		return -1, 0, errors.New("LZF compression is not supported")
	}

	return -1, 0, nil
}

func stringEncoding(data []byte) (str string, bytesConsumed int, err error) {
	size, bytesConsumed, err := parseSizeEncoding(data)
	if err != nil {
		return "", 0, err
	}

	totalBytesConsumed := bytesConsumed + size

	return string(data[bytesConsumed:totalBytesConsumed]), totalBytesConsumed, nil
}
