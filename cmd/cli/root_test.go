package main

import (
	"bytes"
	"flag"
	"log"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/urfave/cli/v2"
)

func TestRootActionAppliesScratchFlagsToRuntimeOnly(t *testing.T) {
	set := flag.NewFlagSet("root", flag.ContinueOnError)
	set.Int("port", 8192, "")
	set.Bool("allow", false, "")
	set.Int("max", 0, "")
	set.String("road", "/", "")
	set.Bool("secure", false, "")
	set.Int("scratch-max-items", 0, "")
	set.String("scratch-max-item-size", "", "")
	if err := set.Set("scratch-max-items", "77"); err != nil {
		t.Fatalf("set scratch-max-items: %v", err)
	}
	if err := set.Set("scratch-max-item-size", "7MB"); err != nil {
		t.Fatalf("set scratch-max-item-size: %v", err)
	}
	ctx := cli.NewContext(cli.NewApp(), set, nil)
	rt := runtime.New(runtime.Snapshot{ScratchMaxItems: 500, ScratchMaxItemSize: "10MB", ScratchMaxItemBytes: 10 * 1024 * 1024})

	if err := rootAction(ctx, rt); err != nil {
		t.Fatalf("rootAction returned error: %v", err)
	}
	snapshot := rt.Snapshot()
	if snapshot.ScratchMaxItems != 77 || snapshot.ScratchMaxItemSize != "7MB" || snapshot.ScratchMaxItemBytes != 7*1024*1024 {
		t.Fatalf("scratch runtime = %+v", snapshot)
	}
}

func TestPrintInfoAlwaysShowsScratchURLWithFallbackHost(t *testing.T) {
	var buf bytes.Buffer
	oldLogWriter := log.Writer()
	oldColorOutput := color.Output
	oldNoColor := color.NoColor
	log.SetOutput(&buf)
	color.Output = &buf
	color.NoColor = true
	t.Cleanup(func() {
		log.SetOutput(oldLogWriter)
		color.Output = oldColorOutput
		color.NoColor = oldNoColor
	})

	printInfo(runtime.Snapshot{
		Port:               8192,
		IP:                 "",
		RootPath:           "/tmp/root",
		ScratchMaxItems:    500,
		ScratchMaxItemSize: "10MB",
	})

	output := buf.String()
	if !strings.Contains(output, "/scratch/") {
		t.Fatalf("printInfo output missing scratch path: %q", output)
	}
	if !strings.Contains(output, "127.0.0.1:8192") {
		t.Fatalf("printInfo output missing fallback host: %q", output)
	}
}
