package main

import (
	"flag"
	"testing"

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
