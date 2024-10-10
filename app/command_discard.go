package main

func (s *server) handleCommandDiscard(client *Client) error {
	if !client.Transaction.IsOpen() {
		_, err := client.Write(respAsError("DISCARD without MULTI"))
		if err != nil {
			return err
		}
		return nil
	}

	err := client.Transaction.Close()
	if err != nil {
		return err
	}

	_, err = client.Write(okSimpleString())
	if err != nil {
		return err
	}

	return nil
}
