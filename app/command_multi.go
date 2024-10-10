package main

func (s *server) handleCommandMulti(client *Client) error {
	err := client.Transaction.Open()
	if err != nil {
		return err
	}

	_, err = client.Write(okSimpleString())
	if err != nil {
		return err
	}
	return nil
}
