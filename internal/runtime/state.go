package runtime

import (
	"io"
	"os"
	stdruntime "runtime"
	"sync"

	"github.com/fatih/color"
	"github.com/tcp404/eutil"
)

const ContextKey = "runtimeSnapshot"

type Snapshot struct {
	RootPath      string
	Port          int
	MaxLevel      uint8
	IsAllowUpload bool
	IsSecure      bool
	IP            string
	Username      string
	PasswordHash  string
	SessionVal    string
	Output        io.Writer
	OS            string
	Pwd           string
}

type PersistentConfig struct {
	RootPath      string
	Port          int
	MaxLevel      uint8
	IsAllowUpload bool
	IsSecure      bool
	Username      string
	PasswordHash  string
}

type Patch struct {
	RootPath      *string
	Port          *int
	MaxLevel      *uint8
	IsAllowUpload *bool
	IsSecure      *bool
	Username      *string
	PasswordHash  *string
	SessionVal    *string
}

type Process struct {
	Output     io.Writer
	OS         string
	IP         string
	Pwd        string
	SessionVal string
}

type Runtime struct {
	mu       sync.RWMutex
	snapshot Snapshot
}

func New(snapshot Snapshot) *Runtime {
	return &Runtime{snapshot: snapshot}
}

func NewProcess() Process {
	pwd, _ := os.Getwd()
	ip, _ := eutil.GetLocalIP()
	return Process{
		Output:     eutil.If[io.Writer](stdruntime.GOOS == "windows", color.Output, os.Stderr),
		OS:         stdruntime.GOOS,
		IP:         ip,
		Pwd:        pwd,
		SessionVal: eutil.RandString(64),
	}
}

func SnapshotFromConfig(cfg PersistentConfig, process Process) Snapshot {
	return Snapshot{
		RootPath:      cfg.RootPath,
		MaxLevel:      cfg.MaxLevel,
		Port:          cfg.Port,
		IsAllowUpload: cfg.IsAllowUpload,
		IsSecure:      cfg.IsSecure,
		Username:      cfg.Username,
		PasswordHash:  cfg.PasswordHash,
		Output:        process.Output,
		OS:            process.OS,
		IP:            process.IP,
		Pwd:           process.Pwd,
		SessionVal:    process.SessionVal,
	}
}

func ProcessFromSnapshot(snapshot Snapshot) Process {
	return Process{
		Output:     snapshot.Output,
		OS:         snapshot.OS,
		IP:         snapshot.IP,
		Pwd:        snapshot.Pwd,
		SessionVal: snapshot.SessionVal,
	}
}

func (r *Runtime) Snapshot() Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.snapshot
}

func (r *Runtime) Update(patch Patch) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if patch.RootPath != nil {
		r.snapshot.RootPath = *patch.RootPath
	}
	if patch.Port != nil {
		r.snapshot.Port = *patch.Port
	}
	if patch.MaxLevel != nil {
		r.snapshot.MaxLevel = *patch.MaxLevel
	}
	if patch.IsAllowUpload != nil {
		r.snapshot.IsAllowUpload = *patch.IsAllowUpload
	}
	if patch.IsSecure != nil {
		r.snapshot.IsSecure = *patch.IsSecure
	}
	if patch.Username != nil {
		r.snapshot.Username = *patch.Username
	}
	if patch.PasswordHash != nil {
		r.snapshot.PasswordHash = *patch.PasswordHash
	}
	if patch.SessionVal != nil {
		r.snapshot.SessionVal = *patch.SessionVal
	}
}
