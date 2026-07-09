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
	"github.com/tcp404/OneTiny/internal/scratch"
)

var (
	ErrServerAlreadyRunning = errors.New("server already running")
	ErrServerNotRunning     = errors.New("server not running")
	ErrRuntimeRequired      = errors.New("runtime config required")
)

type Dependencies struct {
	Runtime   *runtime.Runtime
	AccessLog *accesslog.Logger
	Scratch   *scratch.Store
}

type Manager struct {
	mu       sync.Mutex
	cfg      *runtime.Runtime
	logger   *accesslog.Logger
	scratch  *scratch.Store
	srv      *http.Server
	listener net.Listener
	done     chan error
	stopping bool
}

func NewManager(cfg *runtime.Runtime) *Manager {
	return NewManagerWithDependencies(Dependencies{Runtime: cfg})
}

func NewManagerWithDependencies(deps Dependencies) *Manager {
	scratchStore := deps.Scratch
	if scratchStore == nil && deps.Runtime != nil {
		if store, err := newScratchStore(deps.Runtime.Snapshot()); err == nil {
			scratchStore = store
		}
	}
	return &Manager{cfg: deps.Runtime, logger: deps.AccessLog, scratch: scratchStore}
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

	snapshot := m.cfg.Snapshot()
	if err := m.updateScratchLimits(snapshot); err != nil {
		return err
	}

	srv, listener, err := prepareServer(snapshot, m.cfg, m.logger, m.scratch)
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

	limits := scratchLimitsFromSnapshot(snapshot)
	if err := validateScratchLimits(limits); err != nil {
		return err
	}

	m.mu.Lock()
	if m.stopping {
		m.mu.Unlock()
		return ErrServerAlreadyRunning
	}
	nextScratch := m.scratch
	if nextScratch == nil {
		var err error
		nextScratch, err = scratch.NewStore(limits)
		if err != nil {
			m.mu.Unlock()
			return err
		}
	}
	m.mu.Unlock()

	nextSrv, nextListener, err := prepareServer(snapshot, m.cfg, m.logger, nextScratch)
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
	if err := m.setScratchLimits(limits, nextScratch); err != nil {
		m.mu.Unlock()
		return err
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

	m.mu.Lock()
	defer m.mu.Unlock()
	nextSnapshot := snapshotWithPatch(m.cfg.Snapshot(), patch)
	scratchStore, limits, err := m.prepareScratchStore(nextSnapshot)
	if err != nil {
		return err
	}
	if err := m.setScratchLimits(limits, scratchStore); err != nil {
		return err
	}
	m.cfg.Update(patch)
	return nil
}

func (m *Manager) Config() *runtime.Runtime {
	return m.cfg
}

func (m *Manager) Scratch() *scratch.Store {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.scratch
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

func prepareServer(snapshot runtime.Snapshot, rt *runtime.Runtime, logger *accesslog.Logger, scratchStore *scratch.Store) (*http.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(snapshot.Port))
	if err != nil {
		return nil, nil, fmt.Errorf("start server: %w", err)
	}
	srv, err := defaultBuildHTTPServer(listener, rt, logger, scratchStore)
	if err != nil {
		_ = listener.Close()
		return nil, nil, err
	}
	return srv, listener, nil
}

func defaultBuildHTTPServer(listener net.Listener, rt *runtime.Runtime, logger *accesslog.Logger, scratchStore *scratch.Store) (*http.Server, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	if err := setupEngine(r, rt, logger, scratchStore); err != nil {
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
		RootPath:            &snapshot.RootPath,
		Port:                &snapshot.Port,
		MaxLevel:            &snapshot.MaxLevel,
		IsAllowUpload:       &snapshot.IsAllowUpload,
		IsSecure:            &snapshot.IsSecure,
		Username:            &snapshot.Username,
		PasswordHash:        &snapshot.PasswordHash,
		SessionVal:          &snapshot.SessionVal,
		ScratchMaxItems:     &snapshot.ScratchMaxItems,
		ScratchMaxItemSize:  &snapshot.ScratchMaxItemSize,
		ScratchMaxItemBytes: &snapshot.ScratchMaxItemBytes,
	})
}

func (m *Manager) updateScratchLimits(snapshot runtime.Snapshot) error {
	scratchStore, limits, err := m.prepareScratchStore(snapshot)
	if err != nil {
		return err
	}
	return m.setScratchLimits(limits, scratchStore)
}

func (m *Manager) prepareScratchStore(snapshot runtime.Snapshot) (*scratch.Store, scratch.Limits, error) {
	limits := scratchLimitsFromSnapshot(snapshot)
	if m.scratch != nil {
		if err := validateScratchLimits(limits); err != nil {
			return nil, scratch.Limits{}, err
		}
		return m.scratch, limits, nil
	}
	scratchStore, err := scratch.NewStore(limits)
	if err != nil {
		return nil, scratch.Limits{}, err
	}
	return scratchStore, limits, nil
}

func (m *Manager) setScratchLimits(limits scratch.Limits, prepared *scratch.Store) error {
	if m.scratch == nil {
		if prepared == nil {
			var err error
			prepared, err = scratch.NewStore(limits)
			if err != nil {
				return err
			}
		}
		if err := prepared.UpdateLimits(limits); err != nil {
			return err
		}
		m.scratch = prepared
		return nil
	}
	if err := m.scratch.UpdateLimits(limits); err != nil {
		return err
	}
	return nil
}

func newScratchStore(snapshot runtime.Snapshot) (*scratch.Store, error) {
	return scratch.NewStore(scratchLimitsFromSnapshot(snapshot))
}

func scratchLimitsFromSnapshot(snapshot runtime.Snapshot) scratch.Limits {
	return scratch.Limits{
		MaxItems:     snapshot.ScratchMaxItems,
		MaxItemBytes: int(snapshot.ScratchMaxItemBytes),
	}
}

func validateScratchLimits(limits scratch.Limits) error {
	_, err := scratch.NewStore(limits)
	return err
}

func snapshotWithPatch(snapshot runtime.Snapshot, patch runtime.Patch) runtime.Snapshot {
	if patch.RootPath != nil {
		snapshot.RootPath = *patch.RootPath
	}
	if patch.Port != nil {
		snapshot.Port = *patch.Port
	}
	if patch.MaxLevel != nil {
		snapshot.MaxLevel = *patch.MaxLevel
	}
	if patch.IsAllowUpload != nil {
		snapshot.IsAllowUpload = *patch.IsAllowUpload
	}
	if patch.IsSecure != nil {
		snapshot.IsSecure = *patch.IsSecure
	}
	if patch.Username != nil {
		snapshot.Username = *patch.Username
	}
	if patch.PasswordHash != nil {
		snapshot.PasswordHash = *patch.PasswordHash
	}
	if patch.SessionVal != nil {
		snapshot.SessionVal = *patch.SessionVal
	}
	if patch.ScratchMaxItems != nil {
		snapshot.ScratchMaxItems = *patch.ScratchMaxItems
	}
	if patch.ScratchMaxItemSize != nil {
		snapshot.ScratchMaxItemSize = *patch.ScratchMaxItemSize
	}
	if patch.ScratchMaxItemBytes != nil {
		snapshot.ScratchMaxItemBytes = *patch.ScratchMaxItemBytes
	}
	return snapshot
}
