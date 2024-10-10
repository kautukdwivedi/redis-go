package main

import (
	"encoding/base64"
	"fmt"
)

func (s *server) handleCommandPsync(client *Client) ([]byte, error) {
	resp := fmt.Sprintf("FULLRESYNC %s %d", s.masterReplId, s.masterReplOffset)

	_, err := client.Write(respAsSimpleString(resp))
	if err != nil {
		return nil, err
	}

	emptyFileBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	fileData, err := base64.StdEncoding.DecodeString(emptyFileBase64)
	if err != nil {
		return nil, err
	}

	_, err = client.Write(respAsFileData(fileData))
	if err != nil {
		return nil, err
	}

	fmt.Println("Adding slave...")
	s.slavesMu.Lock()
	s.slaves = append(s.slaves, client.Conn)
	s.slavesMu.Unlock()

	s.dataMu.RLock()
	for k, v := range s.data {
		resp := make([]string, 0, 5)
		resp = append(resp, "SET")
		resp = append(resp, k)
		resp = append(resp, v.Val)
		if px := v.ExpiresIn; px > 0 {
			pxBytes, err := intToByteSlice(px)
			if err != nil {
				continue
			}
			resp = append(resp, "px")
			resp = append(resp, string(pxBytes))
		}
		r, err := respAsArray(resp)
		if err != nil {
			continue
		}

		_, err = client.Write(r)
		if err != nil {
			return nil, err
		}
	}
	s.dataMu.RUnlock()

	return nil, nil
}
