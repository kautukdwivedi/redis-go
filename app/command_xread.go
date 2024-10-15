package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (s *server) handleCommandXREAD(args []string) ([]byte, error) {
	var blockingMillis *int
	streamsArgIdx := -1

	for idx, arg := range args {
		lowerArg := strings.ToLower(arg)

		if lowerArg == "block" {
			blockingMillisInt, err := strconv.Atoi(args[idx+1])
			if err != nil {
				return nil, err
			}

			blockingMillis = &blockingMillisInt
		} else if strings.ToLower(arg) == "streams" {
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

	var createdAfter *time.Time

	if blockingMillis != nil {
		now := time.Now().UTC()
		createdAfter = &now
		if *blockingMillis > 0 {
			timer := time.After(time.Duration(*blockingMillis) * time.Millisecond)
			<-timer
		}
	}

	streamBytesResp := make([][]byte, 0, len(streamKeyAndIds))

	for _, streamKeyAndId := range streamKeyAndIds {
		stream := s.findStream(streamKeyAndId.key)
		if stream == nil {
			fmt.Println("No streams found for key: ", streamKeyAndId.key)
			continue
		}

		if blockingMillis != nil && *blockingMillis == 0 {
			stream.IsBlocking = true
			<-stream.EntryAddedChan
			stream.IsBlocking = false
		}

		entries, err := stream.findEntries(&streamKeyAndId.id, nil)
		if err != nil {
			return nil, err
		}

		if len(entries) == 0 {
			continue
		}

		entriesBytes := make([][]byte, 0, len(entries))

		for _, entry := range entries {
			if createdAfter != nil && entry.Created.Before(*createdAfter) {
				continue
			}

			entryBytesResp, err := entry.encodeToResp()
			if err != nil {
				return nil, err
			}

			entriesBytes = append(entriesBytes, entryBytesResp)
		}

		if len(entriesBytes) == 0 {
			continue
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

	if len(streamBytesResp) == 0 {
		return respAsBulkString(""), nil
	}

	return respAsByteArrays(streamBytesResp)
}

type streamKeyAndId struct {
	key string
	id  string
}
