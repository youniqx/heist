package agentserver

import (
	"context"
	"net/http"
	"time"

	"github.com/youniqx/heist/pkg/erx"
)

// ErrServerError is returned when there was a problem with starting or stopping the Agent HTTP Server.
var ErrServerError = erx.New("Agent Server", "HTTP server failed to operate properly")

const maxStartupTries = 10

func (s *server) ListenAndServer(address string) error {
	log := s.Log.WithValues("address", address)

	errorChannel := make(chan error)
	defer close(errorChannel)

	for i := 0; ; i++ {
		startupLog := log.WithValues("attempt", i)

		if i >= maxStartupTries {
			startupLog.Info("giving up after 10 tries")
			return ErrServerError.WithDetails("failed to start listening, even after 10 retries")
		}

		startupLog.Info("trying to start listening")
		if err := s.StartListening(address, errorChannel); err != nil {
			startupLog.Info("failed to start listening", "error", err)
			time.Sleep(time.Second)
			continue
		}

		startupLog.Info("successfully started listening")

		break
	}

	resultChannel := make(chan error)
	defer close(resultChannel)
	go s.StopOnError(errorChannel, resultChannel)

	workChannel := make(chan bool)
	go s.WorkerLoop(workChannel)

	for range workChannel {
		// Work Channel is closed when the work loop exists
		// This loop is just waiting for the work loop to complete
	}

	err := <-resultChannel
	log.Info("Server has shut down successfully", "error", err)

	return err
}

func (s *server) StopOnError(errorChannel chan error, resultChannel chan error) {
	err := <-errorChannel
	s.Log.Error(err, "Encountered an error while listening, shutting down server")
	s.Stop()
	resultChannel <- err
}

func (s *server) IsListening() bool {
	s.ServerLock.Lock()
	defer s.ServerLock.Unlock()
	return s.Server != nil
}

func (s *server) StartListening(address string, errorChannel chan error) error {
	s.ServerLock.Lock()
	defer s.ServerLock.Unlock()
	if s.Server != nil {
		return ErrServerError.WithDetails("a server is already running, cannot start a second one")
	}

	s.ListenLock.Lock()
	defer s.ListenLock.Unlock()

	s.Server = &http.Server{
		Addr:              address,
		Handler:           s.Mux,
		ReadHeaderTimeout: s.ReadHeaderTimeout,
	}

	go s.Listen(errorChannel)

	return nil
}

func (s *server) Listen(errorChannel chan error) {
	s.ListenLock.Lock()
	defer s.ListenLock.Unlock()
	errorChannel <- s.Server.ListenAndServe()
}

func (s *server) StopListening() error {
	s.ServerLock.Lock()
	defer s.ServerLock.Unlock()

	if s.Server == nil {
		return ErrServerError.WithDetails("server is currently not listening so cannot stop listening")
	}

	if err := s.Server.Shutdown(context.TODO()); err != nil {
		return err
	}

	s.Server = nil

	return nil
}

func (s *server) live(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (s *server) shutdown(writer http.ResponseWriter, request *http.Request) {
	s.Stop()
	writer.WriteHeader(http.StatusOK)
}

func (s *server) ready(writer http.ResponseWriter, request *http.Request) {
	if s.IsSynced() {
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
