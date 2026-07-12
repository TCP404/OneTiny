package main

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type GoBuildOptions struct {
	Kind       Kind
	Target     Target
	Version    string
	OutputPath string
	Package    string
	CGOEnabled string
}

type GoBuildPlan struct {
	Args []string
	Env  map[string]string
}

func NewGoBuildPlan(opts GoBuildOptions) GoBuildPlan {
	env := map[string]string{
		"GOOS":   opts.Target.OS,
		"GOARCH": opts.Target.Arch,
	}
	if opts.CGOEnabled != "" {
		env["CGO_ENABLED"] = opts.CGOEnabled
	}

	return GoBuildPlan{
		Args: []string{
			"build",
			"-trimpath",
			"-ldflags",
			BuildLdflags(opts.Kind, opts.Target, opts.Version),
			"-o",
			opts.OutputPath,
			opts.Package,
		},
		Env: env,
	}
}

func RunGoBuild(opts GoBuildOptions) error {
	if opts.OutputPath == "" {
		return errors.New("go-build requires output path")
	}
	if opts.Package == "" {
		return errors.New("go-build requires package path")
	}

	plan := NewGoBuildPlan(opts)
	cmd := exec.Command("go", plan.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	for key, value := range plan.Env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	return cmd.Run()
}
