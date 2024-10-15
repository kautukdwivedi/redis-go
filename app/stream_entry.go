package main

import "sync"

type StreamEntry struct {
	ID     *StreamEntryId
	Data   map[string]string
	dataMu *sync.Mutex
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

func (e *StreamEntry) AddData(key, val string) {
	e.dataMu.Lock()
	defer e.dataMu.Unlock()

	e.Data[key] = val
}

func (e *StreamEntry) encodeToResp() ([]byte, error) {
	vals := make([]string, 0, len(e.Data)*2)
	for key, val := range e.Data {
		vals = append(vals, key)
		vals = append(vals, val)
	}

	valsResp, err := respAsArray(vals)
	if err != nil {
		return nil, err
	}

	entryBytes := make([][]byte, 0)

	entryBytes = append(entryBytes, respAsBulkString(e.ID.String()))
	entryBytes = append(entryBytes, valsResp)

	return respAsByteArrays(entryBytes)
}
