package rdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	eof              = 0xFF
	selectDB         = 0xFE
	expireTimeSec    = 0xFD
	expireTimeMillis = 0xFC
	resizeDB         = 0xFB
	aux              = 0xFA
)

type value struct {
	Val string
	Exp *time.Time
}

type info struct {
	MagicNumber [5]byte
	Version     [4]byte
}

type RDBFile struct {
	Dir        string
	DBFilename string
	DBs        []*RDBDatabase
	info       info
	auxiliary  map[string]string
}

type RDBDatabase struct {
	Data map[string]value
	mu   *sync.Mutex
}

func NewRDBFile(dir, filename string) *RDBFile {
	return &RDBFile{
		Dir:        dir,
		DBFilename: filename,
		DBs:        make([]*RDBDatabase, 0),
		auxiliary:  make(map[string]string),
	}

}

func (s *RDBFile) Load() error {
	file, err := os.ReadFile(filepath.Join(s.Dir, s.DBFilename))
	if err != nil {
		return err
	}

	in := bytes.NewReader(file)
	err = binary.Read(in, binary.BigEndian, &s.info)

	if err != nil {
		return err
	}

	var databaseHashTableSize int32

outer:
	for {
		fb, errPa := parseByte(in)
		if errPa != nil {
			return errPa

		}

		switch fb {
		case aux:
			key, err := parseString(in)
			if err != nil {
				return err
			}

			val, err := parseString(in)
			if err != nil {
				return err
			}

			s.auxiliary[key] = val
		case selectDB:
			_, _, err = parseLengthEncoding(in)
			if err != nil {
				return err
			}
		case resizeDB:
			databaseHashTableSize, _, err = parseLengthEncoding(in)
			if err != nil {
				return err
			}

			_, _, err = parseLengthEncoding(in)
			if err != nil {
				return err
			}

			db := &RDBDatabase{
				Data: make(map[string]value),
				mu:   &sync.Mutex{},
			}
			s.DBs = append(s.DBs, db)

			for range databaseHashTableSize {
				var valType byte
				var val value

				kpF, err := parseByte(in)
				if err != nil {
					return err
				}

				switch kpF {
				case expireTimeMillis:
					msBuffer := make([]byte, 8)

					err := binary.Read(in, binary.BigEndian, &msBuffer)
					if err != nil {
						return err
					}

					i := int64(msBuffer[0]) + int64(msBuffer[1])<<8 + int64(msBuffer[2])<<16 + int64(msBuffer[3])<<24 + int64(msBuffer[4])<<32 + int64(msBuffer[5])<<40 + int64(msBuffer[6])<<48 + int64(msBuffer[7])<<56
					expiryTime := time.Unix(i/1000, i%1000*1000)
					val.Exp = &expiryTime

					valType, err = parseByte(in)
					if err != nil {
						return err
					}
				case expireTimeSec:
					secBuffer := make([]byte, 4)

					err := binary.Read(in, binary.BigEndian, &secBuffer)
					if err != nil {
						return err
					}

					i := int64(secBuffer[0]) + int64(secBuffer[1])<<8 + int64(secBuffer[2])<<16 + int64(secBuffer[3])<<24
					expiryTime := time.Unix(i, 0)
					val.Exp = &expiryTime

					valType, err = parseByte(in)
					if err != nil {
						return err
					}
				default:
					valType = kpF
				}

				key, err := parseString(in)
				if err != nil {
					return err
				}
				switch valType {
				case 0:
					val.Val, err = parseString(in)
					if err != nil {
						return err
					}
				default:
					return errors.New("type is not implemented")
				}

				db.mu.Lock()
				db.Data[key] = val
				db.mu.Unlock()
			}

		case eof:
			break outer
		default:
			return errors.New("unknown byte marker")
		}
	}

	return nil
}

func parseByte(io io.Reader) (byte, error) {
	var fb byte

	err := binary.Read(io, binary.BigEndian, &fb)
	if err != nil {
		return 0, err
	}

	return fb, nil
}

func parseLengthEncoding(in io.Reader) (int32, byte, error) {
	f, err := parseByte(in)
	if err != nil {
		return 0, 0, err
	}

	encType := f >> 6

	switch encType {
	case 0b00:
		return int32(f & 0x3F), 0, nil
	case 0b01:
		s, err := parseByte(in)
		if err != nil {
			return 0, 0, err
		}

		return int32(f&0x3F) + int32(s), 0, err
	case 0b10:
		var l int32

		err := binary.Read(in, binary.LittleEndian, &l)
		if err != nil {
			return 0, 0, err
		}

		return l, 0, err
	case 0b11:
		return 0, f, nil
	default:
		return 0, 0, errors.New("invalid string encoding")
	}
}

func parseString(in io.Reader) (string, error) {
	length, encoded, errLen := parseLengthEncoding(in)
	if errLen != nil {
		return "", errLen
	}

	if encoded == 0 {
		buf := make([]byte, length)

		err := binary.Read(in, binary.BigEndian, &buf)
		if err != nil {
			return "", err
		}

		return string(buf), err
	}

	switch encoded & byte(0x3F) {
	case 0:
		var l int8

		err := binary.Read(in, binary.LittleEndian, &l)
		if err != nil {
			return "", err
		}

		return strconv.Itoa(int(l)), nil
	case 1:
		var l int16

		err := binary.Read(in, binary.LittleEndian, &l)
		if err != nil {
			return "", err
		}

		return strconv.Itoa(int(l)), nil
	case 2:
		var l int32

		err := binary.Read(in, binary.LittleEndian, &l)
		if err != nil {
			return "", err
		}

		return strconv.Itoa(int(l)), nil
	case 3:
		return "", errors.New("compressed is not support yet")
	default:
		return "", errors.New("invalid string (integer) encoding")
	}
}
