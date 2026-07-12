set dotenv-load := true
set shell := ["bash", "-euo", "pipefail", "-c"]

version := env("VERSION", "v0.6.0")
target := env("TARGET", "")

alias d := dev
alias g := build-gui
alias c := clean
alias gui := build-gui
alias cli := build-cli

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

[doc("Install repository Git hooks")]
[group("Main")]
hooks-install:
    task hooks:install

[doc("Development build: frontend assets, icons, and GUI binary")]
[group("Main")]
dev target=target:
    task build:gui TARGET="{{ target }}" VERSION="{{ version }}"

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

[doc("Produce the installable CLI artifact for TARGET")]
[group("Package")]
package-cli target=target:
    task package:cli TARGET="{{ target }}" VERSION="{{ version }}"

[doc("Produce the installable GUI artifact for TARGET")]
[group("Package")]
package-gui target=target:
    task package:gui TARGET="{{ target }}" VERSION="{{ version }}"

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

[doc("Remove build and dist artifacts")]
[group("Maintenance")]
clean:
    task clean
