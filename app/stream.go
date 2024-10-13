package main

import "sync"

type Stream struct {
	Key     string
	Entries []*StreamEntry
}

type StreamEntry struct {
	ID     string
	Data   map[string]string
	dataMu *sync.Mutex
}

func NewStream(key string) *Stream {
	return &Stream{
		Key:     key,
		Entries: make([]*StreamEntry, 0),
	}
}

func NewStreamEntry(id string) *StreamEntry {
	return &StreamEntry{
		ID:     id,
		Data:   make(map[string]string),
		dataMu: &sync.Mutex{},
	}
}

func (s *Stream) AddEntry(entry *StreamEntry) {
	s.Entries = append(s.Entries, entry)
}

func (e *StreamEntry) AddData(key, val string) {
	e.dataMu.Lock()
	defer e.dataMu.Unlock()

	e.Data[key] = val
}
