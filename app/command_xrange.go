package main

import "fmt"

func (s *server) handleCommandXRANGE(args []string) ([]byte, error) {
	stream := s.findStream(args[0])
	if stream == nil {
		return nil, fmt.Errorf("no streams found for key: %s", args[0])
	}

	start := args[1]
	end := args[2]

	entries, err := stream.findEntries(&start, &end)
	if err != nil {
		return nil, err
	}

	entriesBytes := make([][]byte, 0, len(entries))

	for _, entry := range entries {
		entryBytesResp, err := entry.encodeToResp()
		if err != nil {
			return nil, err
		}

		entriesBytes = append(entriesBytes, entryBytesResp)
	}

	return respAsByteArrays(entriesBytes)
}
