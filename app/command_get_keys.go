package main

func (s *server) handleCommandKeys(client *Client) error {
	resp, err := respAsArray(s.getKeys())
	if err != nil {
		return err
	}

	_, err = client.Write(resp)
	if err != nil {
		return err
	}

	return nil
}
