set dotenv-load := true
set shell := ["bash", "-euo", "pipefail", "-c"]

version := env("VERSION", "v0.6.0")
app_name := "OneTiny"
cli_name := "onetiny-cli"
bin_dir := "build/bin"
frontend_dir := "frontend"
logo_svg := "README/logo.svg"
appicon := "build/appicon.png"
runtime_icon := "internal/gui/assets/appicon.png"
mac_icon := "build/darwin/icons.icns"
windows_icon := "build/windows/icon.ico"
windows_manifest := "internal/gui/assets/windows/wails.exe.manifest"
windows_info := "internal/gui/assets/windows/info.json"
mac_info := "internal/gui/assets/darwin/Info.plist"
gui_main := "./cmd/gui"
cli_main := "./cmd/cli"
upx := env("UPX", "upx")
upx_flags := env("UPX_FLAGS", "--best")
host_goos := if os() == "macos" { "darwin" } else { os() }
target_goos := env("GOOS", host_goos)
goarch := env("GOARCH", `go env GOARCH`)
exe_suffix := if target_goos == "windows" { ".exe" } else { "" }
windows_gui_ldflags := if target_goos == "windows" { " -H windowsgui" } else { "" }
default_build_recipe := if host_goos == "darwin" { if target_goos == "darwin" { "package-mac" } else { "build-gui" } } else { "build-gui" }
go_ldflags := f'-s -w -X github.com/tcp404/OneTiny/internal/constant.VERSION={{version}}{{windows_gui_ldflags}}'
gui_binary := f'{{bin_dir}}/{{app_name}}{{exe_suffix}}'
cli_binary := f'{{bin_dir}}/{{cli_name}}{{exe_suffix}}'
mac_app := f'{{bin_dir}}/{{app_name}}.app'
windows_syso := f'cmd/gui/rsrc_windows_{{goarch}}.syso'

alias b := build
alias d := dev
alias g := build-gui
alias c := clean
alias gui := build-gui
alias cli := build-cli
alias package := package-mac

[default]
[doc("List available recipes")]
[group("Main")]
help:
    @just --list --no-aliases

[doc("Development build: frontend assets, icons, and GUI binary")]
[group("Main")]
dev: build-gui

[doc("Build the default desktop artifact for this host")]
[group("Main")]
build:
    @just {{ default_build_recipe }}

[doc("Build the release artifact for this host")]
[group("Main")]
release:
    #!/usr/bin/env bash
    set -euo pipefail
    just {{ default_build_recipe }}
    if [[ "{{ default_build_recipe }}" != "package-mac" ]]; then
      echo "release: {{ target_goos }} packaging is not implemented; produced {{ gui_binary }}"
    fi

[doc("Show computed build settings")]
[group("Main")]
info:
    @printf "version=%s\n" "{{ version }}"
    @printf "host_goos=%s\n" "{{ host_goos }}"
    @printf "target_goos=%s\n" "{{ target_goos }}"
    @printf "goarch=%s\n" "{{ goarch }}"
    @printf "gui_binary=%s\n" "{{ gui_binary }}"
    @printf "cli_binary=%s\n" "{{ cli_binary }}"
    @printf "default_build_recipe=%s\n" "{{ default_build_recipe }}"

[doc("Build GUI binary into build/bin")]
[group("Build")]
build-gui: _frontend _icons
    #!/usr/bin/env bash
    set -euo pipefail
    if [[ "{{ target_goos }}" == "windows" ]]; then
      just _windows-resource
    fi
    mkdir -p "{{ bin_dir }}"
    go build -ldflags '{{ go_ldflags }}' -o "{{ gui_binary }}" "{{ gui_main }}"

[doc("Build CLI binary into build/bin")]
[group("Build")]
build-cli:
    mkdir -p "{{ bin_dir }}"
    go build -ldflags '{{ go_ldflags }}' -o "{{ cli_binary }}" "{{ cli_main }}"

[doc("macOS: create build/bin/OneTiny.app")]
[group("Package")]
package-mac:
    #!/usr/bin/env bash
    set -euo pipefail
    if [[ "{{ host_goos }}" != "darwin" ]]; then
      echo "package-mac must be run on macOS"
      exit 1
    fi
    if [[ "{{ target_goos }}" != "darwin" ]]; then
      echo "package-mac requires GOOS unset or GOOS=darwin"
      exit 1
    fi
    just build-gui
    rm -rf "{{ mac_app }}"
    mkdir -p "{{ mac_app }}/Contents/MacOS" "{{ mac_app }}/Contents/Resources"
    cp "{{ gui_binary }}" "{{ mac_app }}/Contents/MacOS/{{ app_name }}"
    cp "{{ mac_icon }}" "{{ mac_app }}/Contents/Resources/icons.icns"
    cp "{{ mac_info }}" "{{ mac_app }}/Contents/Info.plist"
    codesign --force --deep --sign - "{{ mac_app }}"

[doc("Windows: package target is not implemented yet")]
[group("Package")]
package-windows:
    @echo "package-windows is not implemented yet; use just build-gui to create {{ gui_binary }}"
    @exit 1

[doc("Linux: package target is not implemented yet")]
[group("Package")]
package-linux:
    @echo "package-linux is not implemented yet; use just build-gui to create {{ gui_binary }}"
    @exit 1

[doc("Compress GUI binary with UPX")]
[group("Maintenance")]
compress: build-gui
    {{ upx }} {{ upx_flags }} "{{ gui_binary }}"

[doc("Remove build artifacts")]
[group("Maintenance")]
clean:
    rm -rf "{{ bin_dir }}"
    rm -f cmd/gui/*.syso

[private]
_frontend:
    npm install --prefix "{{ frontend_dir }}"
    npm run build --prefix "{{ frontend_dir }}"

[private]
_icons: _runtime-icon
    mkdir -p build build/darwin build/windows
    cp "{{ runtime_icon }}" "{{ appicon }}"
    wails3 generate icons -input "{{ appicon }}" -macfilename "{{ mac_icon }}" -windowsfilename "{{ windows_icon }}"

[private]
_runtime-icon:
    #!/usr/bin/env bash
    set -euo pipefail
    if [[ -f "{{ runtime_icon }}" && ! "{{ logo_svg }}" -nt "{{ runtime_icon }}" ]]; then
      exit 0
    fi
    command -v rsvg-convert >/dev/null 2>&1 || {
      echo "rsvg-convert is required to regenerate {{ runtime_icon }} from {{ logo_svg }}"
      exit 1
    }
    mkdir -p internal/gui/assets
    rsvg-convert -w 1024 -h 1024 -o "{{ runtime_icon }}" "{{ logo_svg }}"

[private]
_windows-resource:
    wails3 generate syso -arch "{{ goarch }}" -icon "{{ windows_icon }}" -manifest "{{ windows_manifest }}" -info "{{ windows_info }}" -out "{{ windows_syso }}"
