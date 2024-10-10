package main

func (s *server) handleCommandMulti(client *Client) ([]byte, error) {
	err := client.Transaction.Open()
	if err != nil {
		return nil, err
	}

	return okSimpleString(), nil
}
