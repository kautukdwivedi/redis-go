package main

func (s *server) handleCommandExec(client *Client) ([]byte, error) {
	if client.Transaction.IsOpen() {
		if len(client.Transaction.Queue) == 0 {
			resp, err := respAsArray([]string{})
			if err != nil {
				return nil, err
			}

			_, err = client.Write(resp)
			if err != nil {
				return nil, err
			}
		} else {
			resps := make([][]byte, 0, len(client.Transaction.Queue))
			for _, cmd := range client.Transaction.Queue {
				resp, err := s.handleCommandOnMaster(client, cmd)
				if err != nil {
					return nil, err
				}
				resps = append(resps, resp)
			}
			if len(resps) > 0 {
				respToWrite, err := respAsByteArrays(resps)
				if err != nil {
					return nil, err
				}
				_, err = client.Write(respToWrite)
				if err != nil {
					return nil, err
				}
			}
			go func() {
				for _, cmd := range client.Transaction.Queue {
					if cmd.isWrite {
						s.propagateCommandToSlaves(cmd)
					}
				}
			}()
		}
	}

	err := client.Transaction.Close()
	if err != nil {
		_, err := client.Write(respAsError("EXEC without MULTI"))
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
