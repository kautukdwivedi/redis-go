package main

func (s *server) handleCommandExec(client *Client) error {
	if client.Transaction.IsOpen() {
		if len(client.Transaction.Queue) == 0 {
			resp, err := respAsArray([]string{})
			if err != nil {
				return err
			}

			_, err = client.Write(resp)
			if err != nil {
				return err
			}
		}
	}

	err := client.Transaction.Close()
	if err != nil {
		_, err := client.Write(respAsError("EXEC without MULTI"))
		if err != nil {
			return err
		}
	}

	return nil
}
