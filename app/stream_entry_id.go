package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StreamEntryId struct {
	MillisTime int
	SequenceNr int
}

func (s *Stream) NewStreamEntryId(id string) (*StreamEntryId, error) {
	var millisTime int
	var sequenceNr int

	if id == "*" {
		millisTime = int(time.Now().UTC().UnixMilli())
		sequenceNr = 0
	} else {
		pieces := strings.Split(id, "-")

		if len(pieces) != 2 {
			return nil, ErrInvalidStreamEntrId
		}

		ms, err := strconv.Atoi(pieces[0])
		if err != nil {
			return nil, ErrInvalidStreamEntrId
		}
		millisTime = ms

		if pieces[1] == "*" {
			if s.isEmpty() {
				sequenceNr = 1
			} else {
				matchingEntry := s.findEntryByMillis(millisTime)
				if matchingEntry == nil {
					sequenceNr = 0
				} else {
					sequenceNr = matchingEntry.ID.SequenceNr + 1
				}
			}
		} else {
			nr, err := strconv.Atoi(pieces[1])
			if err != nil {
				return nil, ErrInvalidStreamEntrId
			}
			sequenceNr = nr
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
	}

	return &StreamEntryId{
		MillisTime: millisTime,
		SequenceNr: sequenceNr,
	}, nil
}

func (id *StreamEntryId) String() string {
	return fmt.Sprintf("%d-%d", id.MillisTime, id.SequenceNr)
}
