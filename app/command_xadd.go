package main

import "fmt"

func (s *server) handleCommandXADD(args []string) ([]byte, error) {
	streamKey := args[0]

	stream := s.findStream(streamKey)
	if stream == nil {
		stream = NewStream(streamKey)
		s.streams.Streams = append(s.streams.Streams, stream)
	}

	rawId := args[1]
	entry, err := stream.NewStreamEntry(rawId)
	if err != nil {
		return respAsError(fmt.Sprint("The ID specified in XADD ", err.Error())), nil
	}

	stream.AddEntry(entry)
	if s.streams.IsBlocking {
		s.streams.EntryAddedChan <- true
	}

	for i := 2; i < len(args)-1; i += 2 {
		entry.AddData(args[i], args[i+1])
	}

	return respAsBulkString(entry.ID.String()), nil
}
