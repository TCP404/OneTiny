package updater

import (
	"errors"
	"testing"
)

func TestAssetNameCLIDarwinARM64(t *testing.T) {
	got, err := AssetName(ChannelCLI, Platform{OS: "darwin", Arch: "arm64"})
	if err != nil {
		t.Fatalf("AssetName returned error: %v", err)
	}
	if got != "onetiny-cli-darwin-arm64.zip" {
		t.Fatalf("AssetName = %q, want onetiny-cli-darwin-arm64.zip", got)
	}
}

func TestAssetNameGUIWindowsAMD64(t *testing.T) {
	got, err := AssetName(ChannelGUI, Platform{OS: "windows", Arch: "amd64"})
	if err != nil {
		t.Fatalf("AssetName returned error: %v", err)
	}
	if got != "onetiny-gui-windows-x64.zip" {
		t.Fatalf("AssetName = %q, want onetiny-gui-windows-x64.zip", got)
	}
}

func TestAssetNameRejectsGUILinuxAMD64(t *testing.T) {
	_, err := AssetName(ChannelGUI, Platform{OS: "linux", Arch: "amd64"})
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("AssetName error = %v, want %v", err, ErrUnsupportedPlatform)
	}
}

func TestFindAssetReturnsMatchingReleaseAsset(t *testing.T) {
	release := Release{
		TagName: "v1.2.3",
		Assets: []Asset{
			{Name: "onetiny-cli-linux-x64.zip", DownloadURL: "https://example.com/linux.zip"},
			{Name: "onetiny-cli-darwin-arm64.zip", DownloadURL: "https://example.com/darwin.zip"},
		},
	}

	got, err := FindAsset(release, ChannelCLI, Platform{OS: "darwin", Arch: "aarch64"})
	if err != nil {
		t.Fatalf("FindAsset returned error: %v", err)
	}
	if got.Name != "onetiny-cli-darwin-arm64.zip" {
		t.Fatalf("asset name = %q, want onetiny-cli-darwin-arm64.zip", got.Name)
	}
	if got.DownloadURL != "https://example.com/darwin.zip" {
		t.Fatalf("asset URL = %q, want darwin URL", got.DownloadURL)
	}
}
