package main

func (s *server) handleCommandKeys() ([]byte, error) {
	return respAsArray(s.getKeys())
}
