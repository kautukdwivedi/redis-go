package main

import (
	"encoding/base64"
	"fmt"
	"net"
)


func (s *server) handleCommandPsync(conn net.Conn) error {
	resp := fmt.Sprintf("FULLRESYNC %s %d", s.masterReplId, s.masterReplOffset)

	_, err := conn.Write(respAsSimpleString(resp))
	if err != nil {
		return err
	}

	emptyFileBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	fileData, err := base64.StdEncoding.DecodeString(emptyFileBase64)
	if err != nil {
		return err
	}

	_, err = conn.Write(respAsFileData(fileData))
	if err != nil {
		return err
	}

	fmt.Println("Adding slave...")
	s.slavesMu.Lock()
	s.slaves = append(s.slaves, conn)
	s.slavesMu.Unlock()

	s.dataMu.RLock()
	for k, v := range s.data {
		resp := make([]string, 0, 5)
		resp = append(resp, "SET")
		resp = append(resp, k)
		resp = append(resp, v.val)
		if px := v.expiresIn; px > 0 {
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

		_, err = conn.Write(r)
		if err != nil {
			return err
		}
	}
	s.dataMu.RUnlock()

	return nil
}