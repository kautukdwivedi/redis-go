package main

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrInvalidStreamEntrId                    = errors.New("is invalid")
	ErrStreamEntryIdSmallerThanZero           = errors.New("must be greater than 0-0")
	ErrStreamEntryIdEqualOrSmallerThanTopItem = errors.New("is equal or smaller than the target stream top item")
	ErrInvalidStartIndex                      = errors.New("invalid start index")
	ErrInvalidEndIndex                        = errors.New("invalid end index")
)

type Stream struct {
	Key            string
	Entries        []*StreamEntry
	EntryAddedChan chan bool
	IsBlocking     bool
}

func NewStream(key string) *Stream {
	return &Stream{
		Key:            key,
		Entries:        make([]*StreamEntry, 0),
		EntryAddedChan: make(chan bool),
	}
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

func (s *Stream) findEntryByMillis(millis int) *StreamEntry {
	for _, entry := range s.Entries {
		if entry.ID.MillisTime == millis {
			return entry
		}
	}

	return nil
}

func (s *Stream) findEntries(start, end *string) ([]*StreamEntry, error) {
	if s.isEmpty() {
		return s.Entries, nil
	}

	var startMillis *int
	var startSeqNr *int

	var endMillis *int
	var endSeqNr *int

	updateStartMillis := func() error {
		startMs, startNr, err := parseStreamEntryId(*start)
		if err != nil {
			return err
		}
		startMillis = startMs
		startSeqNr = startNr
		return nil
	}

	updateEndMillis := func() error {
		endMs, endNr, err := parseStreamEntryId(*end)
		if err != nil {
			return err
		}
		endMillis = endMs
		endSeqNr = endNr
		return nil
	}

	if start == nil || *start == "-" {
		err := updateEndMillis()
		if err != nil {
			return nil, err
		}
	} else if end == nil || *end == "+" {
		err := updateStartMillis()
		if err != nil {
			return nil, err
		}
	} else {
		err := updateStartMillis()
		if err != nil {
			return nil, err
		}
		err = updateEndMillis()
		if err != nil {
			return nil, err
		}
	}

	result := make([]*StreamEntry, 0, len(s.Entries))

	for _, entry := range s.Entries {
		entryIdMillis := entry.ID.MillisTime
		entryIdSeqNr := entry.ID.SequenceNr

		startMillisEquality := func() bool {
			if entryIdMillis == *startMillis {
				if entryIdSeqNr >= *startSeqNr {
					result = append(result, entry)
				}
				return true
			}
			return false
		}

		endMillisEquality := func() bool {
			if entryIdMillis == *endMillis {
				if entryIdSeqNr <= *endSeqNr {
					result = append(result, entry)
				}
				return true
			}
			return false
		}

		if startMillis == nil {
			if entryIdMillis < *endMillis {
				result = append(result, entry)
				continue
			}

			endMillisEquality()

			continue
		}

		if endMillis == nil {
			if entryIdMillis > *startMillis {
				result = append(result, entry)
				continue
			}

			startMillisEquality()

			continue
		}

		if startMillis != nil && endMillis != nil {
			if entryIdMillis > *startMillis && entryIdMillis < *endMillis {
				result = append(result, entry)
				continue
			}

			if startMillisEquality() {
				continue
			}

			if endMillisEquality() {
				continue
			}
		}
	}

	return result, nil
}

func parseStreamEntryId(id string) (millis *int, seqNr *int, err error) {
	pieces := strings.Split(id, "-")

	if len(pieces) > 2 {
		return nil, nil, ErrInvalidStreamEntrId
	}

	ms, err := strconv.Atoi(pieces[0])
	if err != nil {
		return nil, nil, err
	}

	var nr *int

	if len(pieces) == 2 {
		n, err := strconv.Atoi(pieces[1])
		if err != nil {
			return nil, nil, err
		}
		nr = &n
	}

	return &ms, nr, nil
}
