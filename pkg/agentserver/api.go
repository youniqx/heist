package agentserver

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/agent"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

var (
	defaultLogger     = controllerruntime.Log.WithName("agent-server")
	readHeaderTimeout = 1 * time.Minute
)

type Server interface {
	http.Handler
	ListenAndServer(address string) error
	IsListening() bool
	IsSynced() bool
	Stop()
}

type server struct {
	Log               logr.Logger
	Agent             agent.Agent
	SyncLock          sync.Mutex
	NextSyncTime      time.Time
	SyncedSecretMap   map[string]string
	StopChannel       chan bool
	ReadHeaderTimeout time.Duration
	StatusLock        sync.Mutex
	GlobalWorkChannel chan bool
	SyncCompleted     bool
	ServerLock        sync.Mutex
	ListenLock        sync.Mutex
	Mux               *http.ServeMux
	Server            *http.Server
}

func (s *server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.Mux.ServeHTTP(writer, request)
}

const stopChannelCapacity = 10

func New(agent agent.Agent) Server {
	instance := &server{
		Log:               defaultLogger,
		Agent:             agent,
		SyncedSecretMap:   make(map[string]string),
		StopChannel:       make(chan bool, stopChannelCapacity),
		Mux:               http.NewServeMux(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	instance.Mux.HandleFunc("/live", instance.live)
	instance.Mux.HandleFunc("/ready", instance.ready)
	instance.Mux.HandleFunc("/shutdown", instance.shutdown)

	return instance
}
