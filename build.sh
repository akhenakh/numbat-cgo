#!/bin/bash
set -euo pipefail

# Requires: cargo-zigbuild (cargo install cargo-zigbuild) and zig,
# plus a nightly toolchain with rust-src for the FreeBSD target:
#   rustup toolchain install nightly --profile minimal --component rust-src

OUT=lib

build() {
  local target="$1" dir="$2"
  shift 2
  echo "==> Building $target -> $OUT/$dir"
  cargo "$@" zigbuild --release --target "$target"
  mkdir -p "$OUT/$dir"
  cp "target/$target/release/libnumbat_cgo.a" "$OUT/$dir/"
}

# Linux (glibc, minimum version supported by zig for wide compatibility)
build x86_64-unknown-linux-gnu  linux_amd64
build aarch64-unknown-linux-gnu linux_arm64
build riscv64gc-unknown-linux-gnu linux_riscv64

# macOS (zig provides the libc stubs, no Apple SDK needed)
build aarch64-apple-darwin darwin_arm64
build x86_64-apple-darwin  darwin_amd64

# Windows (GNU toolchain)
build x86_64-pc-windows-gnu windows_amd64

# FreeBSD is a tier-3 Rust target: rust-std must be built from source
# with a nightly toolchain (-Z build-std).
echo "==> Building aarch64-unknown-freebsd -> $OUT/freebsd_arm64"
cargo +nightly zigbuild --release --target aarch64-unknown-freebsd \
  -Z build-std=std,panic_abort
mkdir -p "$OUT/freebsd_arm64"
cp target/aarch64-unknown-freebsd/release/libnumbat_cgo.a "$OUT/freebsd_arm64/"

echo "==> Done"
