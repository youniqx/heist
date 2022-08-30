package agentserver

func (s *server) Stop() {
	s.StopChannel <- true
	if s.GlobalWorkChannel != nil {
		for range s.GlobalWorkChannel {
			// Wait for server to shut down
		}
	}
}
