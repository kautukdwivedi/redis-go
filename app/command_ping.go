package main

func (s *server) handleCommandPing(client *Client) error {
	if s.isMaster() {
		_, err := client.Write(respAsSimpleString("PONG"))
		if err != nil {
			return err
		}
	}

	return nil
}
