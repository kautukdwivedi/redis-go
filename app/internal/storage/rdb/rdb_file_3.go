package rdb

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"

	"errors"

	"fmt"

	"io"

	"strconv"

	"sync"

	"time"
)

const REDIS_EOF = 0xFF // EOF	End of the RDB file

const REDIS_SELECTDB = 0xFE // SELECTDB	Database Selector

const REDIS_EXPIRETIME = 0xFD // EXPIRETIME	Expire time in seconds, see Key Expiry Timestamp

const REDIS_EXPIRETIMEMS = 0xFC // EXPIRETIMEMS	Expire time in milliseconds, see Key Expiry Timestamp

const REDIS_RESIZEDB = 0xFB // RESIZEDB	Hash table sizes for the main keyspace and expires, see Resizedb information

const REDIS_AUX = 0xFA // AUX	Auxiliary fields. Arbitrary key-value settings, see Auxiliary fields

type value struct {
	Val string

	Exp *time.Time
}

type info struct {
	MagicNumber [5]byte

	Version [4]byte
}

type RDBFile3 struct {
	Dir        string
	DBFilename string
	Data          map[string]value

	mutex sync.RWMutex

	info info

	auxiliary Auxiliary
}

type Auxiliary map[string]string

func NewRDBFile3(dir, filename string) *RDBFile3 {

	return &RDBFile3{
		Dir:        dir,
		DBFilename: filename,
		Data:          make(map[string]value),

		mutex: sync.RWMutex{},

		auxiliary: make(Auxiliary),
	}

}

func (s *RDBFile3) Load() error {
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

	var expiryHashTableSize int32

	var dbNumber int32

	loop := true

	for loop {

		fb, errPa := parseByte(in)

		if errPa != nil {

			return errPa

		}

		switch fb {

		case REDIS_AUX:

			key, err := parseString(in)

			if err != nil {

				return err

			}

			val, err := parseString(in)

			if err != nil {

				return err

			}

			s.auxiliary[key] = val

			break

		case REDIS_SELECTDB:

			dbNumber, _, err = parseLengthEncoding(in)

			if err != nil {

				return err

			}

		case REDIS_RESIZEDB:

			databaseHashTableSize, _, err = parseLengthEncoding(in)

			if err != nil {

				return err

			}

			// todo exp size missed

			_, _, err = parseLengthEncoding(in)

			if err != nil {

				return err

			}

			for range databaseHashTableSize {

				var valType byte

				var val value

				kpF, err := parseByte(in)

				if err != nil {

					return err

				}

				switch kpF {

				case REDIS_EXPIRETIMEMS:

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

				case REDIS_EXPIRETIME:

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

				s.Data[key] = val

			}

		case REDIS_EOF:

			loop = false

		default:

			return errors.New("unknown byte marker")

		}

	}

	fmt.Println(dbNumber)

	fmt.Println(databaseHashTableSize)

	fmt.Println(expiryHashTableSize)

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

// return length as int 32 or first byte if it's encoded in a special format

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

		// The next object is encoded in a special format, return first byte instead of length

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

	// it's not a special format just read next bytes[length]

	if encoded == 0 {

		buf := make([]byte, length)

		err := binary.Read(in, binary.BigEndian, &buf)

		if err != nil {

			return "", err

		}

		return string(buf), err

	}

	// get the special format by remaining 6 bits and read it

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

func (s *RDBFile3) Get(key string) (string, bool) {

	s.mutex.RLock()

	defer s.mutex.RUnlock()

	v, ok := s.Data[key]

	if ok && (v.Exp == nil || time.Now().Before(*v.Exp)) {

		return v.Val, ok

	}

	return "", false

}

func (s *RDBFile3) Set(key string, val string, dur time.Duration) {

	s.mutex.Lock()

	defer s.mutex.Unlock()

	var exp *time.Time

	if dur != 0 {

		n := time.Now().Add(dur)

		exp = &n

	}

	s.Data[key] = value{Val: val, Exp: exp}

}

func (s *RDBFile3) FlushAll() {

	s.mutex.Lock()

	defer s.mutex.Unlock()

	s.Data = make(map[string]value)

}

func (s *RDBFile3) FindBy(pattern string) ([]string, error) {

	res := make([]string, 0)

	for i := range s.Data {

		// @todo just hard code '*' pattern, add regexp support latter

		if i == pattern || pattern == "*" {

			res = append(res, i)

		}

	}

	return res, nil

}

func (s *RDBFile3) MagicNumber() string {

	return string(s.info.MagicNumber[:])

}

func (s *RDBFile3) Version() string {

	return string(s.info.Version[:])

}
