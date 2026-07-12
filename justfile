set dotenv-load := true
set shell := ["bash", "-euo", "pipefail", "-c"]

version := env("VERSION", "v0.6.0")
target := env("TARGET", "")

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

[doc("Show computed build settings")]
[group("Main")]
info target=target:
    task info TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Run baseline checks")]
[group("Main")]
check:
    task check

[doc("Run required checks before commit")]
[group("Main")]
precommit:
    task precommit

[doc("Run required checks before push")]
[group("Main")]
prepush:
    task prepush

[doc("Development build: frontend assets, icons, and GUI binary")]
[group("Main")]
dev target=target:
    task build:gui TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Build the default GUI artifact for this host")]
[group("Main")]
build target=target:
    task build TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Build release GUI artifact for TARGET and update checksums")]
[group("Main")]
release target=target:
    task release TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Build GUI binary into its dist staging directory")]
[group("Build")]
build-gui target=target:
    task build:gui TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Build CLI binary into its dist staging directory")]
[group("Build")]
build-cli target=target:
    task build:cli TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Build frontend assets")]
[group("Build")]
frontend:
    task build:frontend

[doc("Generate Wails icons")]
[group("Build")]
icons:
    task build:icons

[doc("Format Go source files")]
[group("Quality")]
format:
    task format

[doc("Apply Go source modernization fixes")]
[group("Quality")]
fix:
    task fix

[doc("Run Go lint baseline")]
[group("Quality")]
lint:
    task lint

[doc("Run Go tests")]
[group("Quality")]
test:
    task test

[doc("Build and archive CLI release artifact")]
[group("Dist")]
dist-cli target=target:
    task dist:cli TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Build and archive GUI release artifact")]
[group("Dist")]
dist-gui target=target:
    task dist:gui TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Generate release checksums")]
[group("Dist")]
checksums:
    task dist:checksums

[doc("macOS: create and archive build/bin/OneTiny.app")]
[group("Package")]
package-mac target=target:
    task package:mac TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Windows package placeholder")]
[group("Package")]
package-windows:
    task package:windows

[doc("Linux package placeholder")]
[group("Package")]
package-linux:
    task package:linux

[doc("Remove build and dist artifacts")]
[group("Maintenance")]
clean:
    task clean

[private]
_frontend:
    task build:frontend

[private]
_icons:
    task build:icons

[private]
_windows-resource target=target:
    task build:windows-resource TARGET="{{ target }}"
