#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"
[[ "$(uname -s)/$(uname -m)" == Linux/x86_64 ]] || { echo "Build amd64 on Linux/x86_64; use CI matrix for other targets" >&2; exit 2; }
export GOWORK=off GOPROXY=off GOSUMDB=off
OUT="$ROOT/release/bin/linux/amd64"
mkdir -p "$OUT"
go test ./...
go vet ./...
for entry in 'trest-api ./cmd/api' 'trest-worker ./cmd/worker' 'trest ./cmd/trest'; do
 read -r name pkg <<< "$entry"
 CGO_ENABLED=1 go build -trimpath -ldflags='-s -w -buildid=' -o "$OUT/$name" "$pkg"
done
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' -o "$OUT/trestctl" ./cmd/trestctl
(cd bootstrap-installer; go test ./...; CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' -o "$OUT/trest-install" .)
(cd trest-installer; go test ./...; go vet ./...; CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' -o "$OUT/trest-installer" ./cmd/trest)
(cd proektirovka-sdaniy; go test ./...; go vet ./...; CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' -o "$OUT/generate-dxf" ./cmd/generate)
chmod 0755 "$OUT"/*
cp "$OUT/trest-install" bootstrap-installer/bootstrap-installer
mkdir -p release/bin/windows/amd64
(cd bootstrap-installer; GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' -o "$ROOT/release/bin/windows/amd64/trest-install.exe" .)
cp "$OUT/generate-dxf" proektirovka-sdaniy/release/bin/linux/amd64/generate-dxf
sha256sum release/bin/linux/amd64/* release/bin/windows/amd64/* > release/bin/SHA256SUMS
sha256sum -c release/bin/SHA256SUMS
echo 'Binary build completed (this does NOT claim Docker E2E or frontend/Python build).'
