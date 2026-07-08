package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/runtime"
)

var (
	ErrServerAlreadyRunning = errors.New("server already running")
	ErrServerNotRunning     = errors.New("server not running")
	ErrRuntimeRequired      = errors.New("runtime config required")
)

type Dependencies struct {
	Runtime   *runtime.Runtime
	AccessLog *accesslog.Logger
}

type Manager struct {
	mu       sync.Mutex
	cfg      *runtime.Runtime
	logger   *accesslog.Logger
	srv      *http.Server
	listener net.Listener
	done     chan error
	stopping bool
}

func NewManager(cfg *runtime.Runtime) *Manager {
	return &Manager{cfg: cfg}
}

func NewManagerWithDependencies(deps Dependencies) *Manager {
	return &Manager{cfg: deps.Runtime, logger: deps.AccessLog}
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cfg == nil {
		return ErrRuntimeRequired
	}
	if m.srv != nil || m.stopping {
		return ErrServerAlreadyRunning
	}

	srv, listener, err := prepareServer(m.cfg.Snapshot(), m.cfg, m.logger)
	if err != nil {
		return err
	}
	done := make(chan error, 1)
	m.srv = srv
	m.listener = listener
	m.done = done

	go m.serve(srv, listener, done)
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	if m.srv == nil || m.stopping {
		m.mu.Unlock()
		return ErrServerNotRunning
	}
	srv := m.srv
	listener := m.listener
	done := m.done
	m.stopping = true
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	shutdownErr := srv.Shutdown(ctx)
	var serveErr error
	select {
	case serveErr = <-done:
	case <-ctx.Done():
		if listener != nil {
			_ = listener.Close()
		}
		_ = srv.Close()
		serveErr = <-done
	}

	m.mu.Lock()
	if m.srv == srv {
		m.srv = nil
		m.listener = nil
		m.done = nil
	}
	m.stopping = false
	m.mu.Unlock()

	if shutdownErr != nil {
		return shutdownErr
	}
	return serveErr
}

func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil && !errors.Is(err, ErrServerNotRunning) {
		return err
	}
	return m.Start()
}

func (m *Manager) RestartWithSnapshot(snapshot runtime.Snapshot, commit func() error) error {
	if m.cfg == nil {
		return ErrRuntimeRequired
	}

	m.mu.Lock()
	if m.stopping {
		m.mu.Unlock()
		return ErrServerAlreadyRunning
	}
	m.mu.Unlock()

	nextSrv, nextListener, err := prepareServer(snapshot, m.cfg, m.logger)
	if err != nil {
		return err
	}
	nextDone := make(chan error, 1)
	closeNext := true
	defer func() {
		if closeNext {
			_ = nextListener.Close()
			_ = nextSrv.Close()
		}
	}()

	m.mu.Lock()
	if m.stopping {
		m.mu.Unlock()
		return ErrServerAlreadyRunning
	}
	if commit != nil {
		if err := commit(); err != nil {
			m.mu.Unlock()
			return err
		}
	}
	oldSrv := m.srv
	oldListener := m.listener
	oldDone := m.done
	m.srv = nextSrv
	m.listener = nextListener
	m.done = nextDone
	applySnapshot(m.cfg, snapshot)
	m.mu.Unlock()

	closeNext = false
	go m.serve(nextSrv, nextListener, nextDone)
	_ = shutdownPrevious(oldSrv, oldListener, oldDone)
	return nil
}

func (m *Manager) ApplyRuntime(patch runtime.Patch) error {
	if m.cfg == nil {
		return ErrRuntimeRequired
	}
	m.cfg.Update(patch)
	return nil
}

func (m *Manager) Config() *runtime.Runtime {
	return m.cfg
}

func (m *Manager) Done() <-chan error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.done
}

func (m *Manager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.srv != nil && !m.stopping
}

func (m *Manager) Status() runtime.Snapshot {
	if m.cfg == nil {
		return runtime.Snapshot{}
	}
	return m.cfg.Snapshot()
}

func (m *Manager) serve(srv *http.Server, listener net.Listener, done chan<- error) {
	err := srv.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
		err = nil
	}

	m.mu.Lock()
	if m.srv == srv && !m.stopping {
		m.srv = nil
		m.listener = nil
		m.stopping = false
	}
	m.mu.Unlock()

	done <- err
}

func prepareServer(snapshot runtime.Snapshot, rt *runtime.Runtime, logger *accesslog.Logger) (*http.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(snapshot.Port))
	if err != nil {
		return nil, nil, fmt.Errorf("start server: %w", err)
	}
	srv, err := defaultBuildHTTPServer(listener, rt, logger)
	if err != nil {
		_ = listener.Close()
		return nil, nil, err
	}
	return srv, listener, nil
}

func defaultBuildHTTPServer(listener net.Listener, rt *runtime.Runtime, logger *accesslog.Logger) (*http.Server, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	if err := setupEngine(r, rt, logger); err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:    listener.Addr().String(),
		Handler: r,
	}, nil
}

func shutdownPrevious(srv *http.Server, listener net.Listener, done <-chan error) error {
	if srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	shutdownErr := srv.Shutdown(ctx)
	var serveErr error
	if done != nil {
		select {
		case serveErr = <-done:
		case <-ctx.Done():
			if listener != nil {
				_ = listener.Close()
			}
			_ = srv.Close()
			serveErr = <-done
		}
	}
	if shutdownErr != nil {
		return shutdownErr
	}
	return serveErr
}

func applySnapshot(cfg *runtime.Runtime, snapshot runtime.Snapshot) {
	cfg.Update(runtime.Patch{
		RootPath:      &snapshot.RootPath,
		Port:          &snapshot.Port,
		MaxLevel:      &snapshot.MaxLevel,
		IsAllowUpload: &snapshot.IsAllowUpload,
		IsSecure:      &snapshot.IsSecure,
		Username:      &snapshot.Username,
		PasswordHash:  &snapshot.PasswordHash,
		SessionVal:    &snapshot.SessionVal,
	})
}
