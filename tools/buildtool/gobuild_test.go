package main

import (
	"reflect"
	"testing"
)

func TestNewGoBuildPlanSetsTargetEnvAndLdflags(t *testing.T) {
	target, err := ParseTarget("windows-amd64")
	if err != nil {
		t.Fatal(err)
	}

	plan := NewGoBuildPlan(GoBuildOptions{
		Kind:       KindGUI,
		Target:     target,
		Version:    "v0.6.0",
		OutputPath: "dist/onetiny-gui-windows-x64/OneTiny.exe",
		Package:    "./cmd/gui",
		CGOEnabled: "0",
	})

	wantArgs := []string{
		"build",
		"-trimpath",
		"-ldflags",
		"-s -w -H windowsgui -X github.com/tcp404/OneTiny/internal/version.Version=v0.6.0",
		"-o",
		"dist/onetiny-gui-windows-x64/OneTiny.exe",
		"./cmd/gui",
	}
	if !reflect.DeepEqual(plan.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", plan.Args, wantArgs)
	}

	wantEnv := map[string]string{
		"GOOS":        "windows",
		"GOARCH":      "amd64",
		"CGO_ENABLED": "0",
	}
	if !reflect.DeepEqual(plan.Env, wantEnv) {
		t.Fatalf("Env = %#v, want %#v", plan.Env, wantEnv)
	}
}
