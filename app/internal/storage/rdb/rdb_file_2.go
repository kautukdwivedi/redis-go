package rdb

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type RedisObject struct {
	Key      string
	Value    string
	ExpiryAt time.Time
}

type RDBFile2 struct {
	Dir        string
	DBFilename string
	Data       []*RedisObject
	reader     *bufio.Reader
	version    int
	db         int
}

func NewRDBFile2(dir, filename string) *RDBFile2 {
	return &RDBFile2{
		Dir:        dir,
		DBFilename: filename,
		Data:       make([]*RedisObject, 0),
	}
}

func (rf *RDBFile2) Parse() error {
	file, err := os.Open(filepath.Join(rf.Dir, rf.DBFilename))
	if err != nil {
		return fmt.Errorf("error opening rdb file: %s", err)
	}
	defer file.Close()

	rf.reader = bufio.NewReader(file)

	err = rf.readHeader()
	if err != nil {
		return err
	}

	return rf.parseInternal()
}

func (rf *RDBFile2) readHeader() error {
	signature := make([]byte, 5)
	if _, err := io.ReadFull(rf.reader, signature); err != nil {
		return err
	}
	if string(signature) != "REDIS" {
		return fmt.Errorf("invalid RDB file signature")
	}

	versionBytes := make([]byte, 4)
	if _, err := io.ReadFull(rf.reader, versionBytes); err != nil {
		return err
	}

	versionStr := string(versionBytes)

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return fmt.Errorf("invalid db version: %s", versionStr)
	}

	rf.version = version

	return nil
}

func (rf *RDBFile2) parseInternal() error {
	for {
		opcode, err := rf.reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch opcode {
		case 0xFA:
			if err := rf.skipAuxField(); err != nil {
				return err
			}
		case 0xFB:
			if err := rf.skipResizeDB(); err != nil {
				fmt.Println("Got error here!!!, ", err)
				return err
			}
		case 0xFF:
			return nil
		case 0xFE:
			rf.db++
			fmt.Printf("Switching to db: %d\n", rf.db)
		case 0xFC:
			expireTime, err := rf.readMillisecondsTime()
			if err != nil {
				return err
			}
			if err := rf.parseKeyValuePair(expireTime); err != nil {
				return err
			}
		case 0xFD:
			expireTime, err := rf.readSecondsTime()
			if err != nil {
				return err
			}
			if err := rf.parseKeyValuePair(expireTime); err != nil {
				return err
			}
		default:
			if err := rf.parseKeyValuePair(time.Now().UTC()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rf *RDBFile2) skipAuxField() error {
	if err := rf.skipString(); err != nil {
		return err
	}
	return rf.skipString()
}

func (rf *RDBFile2) skipResizeDB() error {
	if err := rf.skipLength(); err != nil {
		return err
	}
	return rf.skipLength()
}

func (rf *RDBFile2) skipString() error {
	length, err := rf.readLength()
	if err != nil {
		return err
	}

	_, err = rf.reader.Discard(int(length))
	return err
}

func (rf *RDBFile2) skipLength() error {
	_, err := rf.readLength()
	return err
}

func (rf *RDBFile2) skipObject(valueType byte) error {
	switch valueType {
	case 1, 2, 3:
		length, err := rf.readLength()
		if err != nil {
			return err
		}
		for i := uint64(0); i < length; i++ {
			if err := rf.skipString(); err != nil {
				return err
			}
		}
	case 4:
		length, err := rf.readLength()
		if err != nil {
			return err
		}
		for i := uint64(0); i < length; i++ {
			if err := rf.skipString(); err != nil {
				return err
			}
			if err := rf.skipString(); err != nil {
				return err
			}
		}
	case 9, 10, 11, 12, 13:
		return rf.skipString()
	default:
		fmt.Println("unknown value type: ", valueType)
		return rf.skipString()
	}
	return nil
}

func (rf *RDBFile2) readLength() (uint64, error) {
	b, err := rf.reader.ReadByte()
	if err != nil {
		return 0, err
	}

	switch b >> 6 {
	case 0:
		return uint64(b & 0x3F), nil
	case 1:
		next, err := rf.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint64(((b & 0x3F) << 8)) | uint64(next), nil
	case 2:
		buf := make([]byte, 4)
		if _, err := io.ReadFull(rf.reader, buf); err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint32(buf)), nil
	case 3:
		switch b & 0x3F {
		case 0:
			buf := make([]byte, 8)
			if _, err := io.ReadFull(rf.reader, buf); err != nil {
				return 0, err
			}
			return binary.LittleEndian.Uint64(buf), nil
		case 1:
			buf := make([]byte, 2)
			if _, err := io.ReadFull(rf.reader, buf); err != nil {
				return 0, err
			}
			return uint64(int16(binary.LittleEndian.Uint16(buf))), nil
		case 2:
			buf := make([]byte, 4)
			if _, err := io.ReadFull(rf.reader, buf); err != nil {
				return 0, err
			}
			return uint64(int32(binary.LittleEndian.Uint32(buf))), nil
		default:
			return 0, fmt.Errorf("unknown special encoding: %d", b&0x3F)
		}
	default:
		return 0, fmt.Errorf("invalid length encoding")
	}
}

func (rf *RDBFile2) readString() (string, error) {
	length, err := rf.readLength()
	if err != nil {
		return "", err
	}

	str := make([]byte, length)
	if _, err := io.ReadFull(rf.reader, str); err != nil {
		return "", err
	}

	return string(str), nil
}

func (rf *RDBFile2) readMillisecondsTime() (time.Time, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(rf.reader, buf); err != nil {
		return time.Time{}, err
	}
	millis := binary.LittleEndian.Uint64(buf)
	return time.Unix(int64(millis/1000), int64(millis%1000)*int64(time.Millisecond)), nil
}

func (rf *RDBFile2) readSecondsTime() (time.Time, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(rf.reader, buf); err != nil {
		return time.Time{}, err
	}
	seconds := binary.LittleEndian.Uint32(buf)
	return time.Unix(int64(seconds), 0), nil
}

func (rf *RDBFile2) parseKeyValuePair(expireTime time.Time) error {
	key, err := rf.readString()
	if err != nil {
		return err
	}

	valueType, err := rf.reader.ReadByte()
	if err != nil {
		return err
	}

	if valueType == 0 {
		value, err := rf.readString()
		if err != nil {
			return err
		}

		rf.Data = append(rf.Data, &RedisObject{
			Key:      key,
			Value:    value,
			ExpiryAt: expireTime,
		})
	} else {
		if err := rf.skipObject(valueType); err != nil {
			return err
		}
	}

	return nil
}
