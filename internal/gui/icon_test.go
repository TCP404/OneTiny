package gui

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/tcp404/OneTiny/resource"
)

func TestAppIconMatchesBrandLogo(t *testing.T) {
	want, err := resource.FS.ReadFile("logo/logo.png")
	if err != nil {
		t.Fatalf("read canonical logo: %v", err)
	}

	if !bytes.Equal(appIcon, want) {
		t.Fatal("app icon does not match resource/logo/logo.png; run `just _icons` or `make icons`")
	}
}

func TestAppIconIsPNG(t *testing.T) {
	if _, err := png.Decode(bytes.NewReader(appIcon)); err != nil {
		t.Fatalf("decode app icon png: %v", err)
	}
}
