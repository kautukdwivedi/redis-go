package main

func (s *server) handleCommandPing(client *Client) ([]byte, error) {
	if s.isMaster() {
		return respAsSimpleString("PONG"), nil
	}

	return nil, nil
}
