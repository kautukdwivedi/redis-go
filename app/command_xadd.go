package main

func (s *server) handleCommandXADD(args []string) ([]byte, error) {
	streamKey := args[0]

	stream := s.findStream(streamKey)
	if stream == nil {
		stream = NewStream(streamKey)
		s.streams = append(s.streams, stream)
	}

	entryId := args[1]
	entry := NewStreamEntry(entryId)
	stream.AddEntry(entry)

	for i := 2; i < len(args)-1; i += 2 {
		entry.AddData(args[i], args[i+1])
	}

	return respAsBulkString(entryId), nil
}
