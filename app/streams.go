package main

type Streams struct {
	Streams        []*Stream
	EntryAddedChan chan bool
	IsBlocking     bool
}

func NewStreams() *Streams {
	return &Streams{
		Streams:        make([]*Stream, 0),
		EntryAddedChan: make(chan bool),
	}
}

func (s *Streams) blockAndWaitForEntryAdded() {
	s.IsBlocking = true
	<-s.EntryAddedChan
	s.IsBlocking = false
}
