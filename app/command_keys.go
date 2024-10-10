package main

func (s *server) handleCommandKeys(client *Client) ([]byte, error) {
	return respAsArray(s.getKeys())
}
