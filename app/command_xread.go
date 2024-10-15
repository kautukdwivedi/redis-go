package main

import (
	"errors"
	"fmt"
	"strings"
)

func (s *server) handleCommandXREAD(args []string) ([]byte, error) {
	streamsArgIdx := -1
	for idx, arg := range args {
		if strings.ToLower(arg) == "streams" {
			streamsArgIdx = idx
			break
		}
	}

	if streamsArgIdx == -1 {
		return nil, errors.New("invalid XREAD command - \"streams\" argument not found")
	}

	streamArgs := args[streamsArgIdx+1:]

	streamKeyAndIds := make([]streamKeyAndId, 0, len(streamArgs)/2)

	i := 0
	j := len(streamArgs) / 2

	for j < len(streamArgs) {
		streamKeyAndIds = append(streamKeyAndIds, streamKeyAndId{
			key: streamArgs[i],
			id:  streamArgs[j],
		})
		i++
		j++
	}

	if len(streamKeyAndIds) == 0 {
		return nil, errors.New("no streams to read from")
	}

	streamBytesResp := make([][]byte, 0, len(streamKeyAndIds))

	for _, streamKeyAndId := range streamKeyAndIds {
		stream := s.findStream(streamKeyAndId.key)
		if stream == nil {
			fmt.Println("No streams found for key: ", streamKeyAndId.key)
			continue
		}

		entries, err := stream.findEntries(&streamKeyAndId.id, nil)
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

		streamBytes := make([][]byte, 0, 2)

		entriesBytesResp, err := respAsByteArrays(entriesBytes)
		if err != nil {
			return nil, err
		}

		streamBytes = append(streamBytes, respAsBulkString(streamKeyAndId.key))
		streamBytes = append(streamBytes, entriesBytesResp)

		encodedStreamBytes, err := respAsByteArrays(streamBytes)
		if err != nil {
			return nil, err
		}

		streamBytesResp = append(streamBytesResp, encodedStreamBytes)
	}

	return respAsByteArrays(streamBytesResp)
}

type streamKeyAndId struct {
	key string
	id  string
}
