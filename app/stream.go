package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrInvalidStreamEntrId                    = errors.New("is invalid")
	ErrStreamEntryIdSmallerThanZero           = errors.New("must be greater than 0-0")
	ErrStreamEntryIdEqualOrSmallerThanTopItem = errors.New("is equal or smaller than the target stream top item")
)

type Stream struct {
	Key     string
	Entries []*StreamEntry
}

type StreamEntry struct {
	ID     *StreamEntryId
	Data   map[string]string
	dataMu *sync.Mutex
}

type StreamEntryId struct {
	MillisTime int
	SequenceNr int
}

func NewStream(key string) *Stream {
	return &Stream{
		Key:     key,
		Entries: make([]*StreamEntry, 0),
	}
}

func (s *Stream) NewStreamEntry(id string) (*StreamEntry, error) {
	entryId, err := s.NewStreamEntryId(id)
	if err != nil {
		return nil, err
	}

	return &StreamEntry{
		ID:     entryId,
		Data:   make(map[string]string),
		dataMu: &sync.Mutex{},
	}, nil
}

func (s *Stream) NewStreamEntryId(id string) (*StreamEntryId, error) {
	pieces := strings.Split(id, "-")

	if len(pieces) != 2 {
		return nil, ErrInvalidStreamEntrId
	}

	millisTime, err := strconv.Atoi(pieces[0])
	if err != nil {
		return nil, ErrInvalidStreamEntrId
	}

	sequenceNr, err := strconv.Atoi(pieces[1])
	if err != nil {
		return nil, ErrInvalidStreamEntrId
	}

	if millisTime <= 0 && sequenceNr <= 0 {
		return nil, ErrStreamEntryIdSmallerThanZero
	}

	if !s.isEmpty() {
		lastEntryId := s.lastEntry().ID

		if lastEntryId.MillisTime > millisTime {
			return nil, ErrStreamEntryIdEqualOrSmallerThanTopItem
		}

		if lastEntryId.MillisTime == millisTime && lastEntryId.SequenceNr >= sequenceNr {
			return nil, ErrStreamEntryIdEqualOrSmallerThanTopItem
		}
	}

	return &StreamEntryId{
		MillisTime: millisTime,
		SequenceNr: sequenceNr,
	}, nil
}

func (s *Stream) AddEntry(entry *StreamEntry) {
	s.Entries = append(s.Entries, entry)
}

func (s *Stream) isEmpty() bool {
	return len(s.Entries) == 0
}

func (s *Stream) lastEntry() *StreamEntry {
	return s.Entries[len(s.Entries)-1]
}

func (e *StreamEntry) AddData(key, val string) {
	e.dataMu.Lock()
	defer e.dataMu.Unlock()

	e.Data[key] = val
}

func (id *StreamEntryId) String() string {
	return fmt.Sprintf("%d-%d", id.MillisTime, id.SequenceNr)
}
