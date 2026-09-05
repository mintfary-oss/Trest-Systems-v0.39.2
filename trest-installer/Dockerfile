# Builds the trest-installer orchestrator binary. The runtime image also
# needs `git` (to clone/update the tracked repo) and the `docker` CLI (to
# start the magasin-777 and super-sistema compose projects) — run the
# resulting container with the host's docker socket mounted, e.g.:
#
#   docker run --rm \
#     -v /var/run/docker.sock:/var/run/docker.sock \
#     -v trest-work:/work \
#     -p 9091:9091 \
#     trest-installer audit

FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/trest-installer ./cmd/trest

FROM alpine:3.22
RUN apk add --no-cache ca-certificates git docker-cli docker-cli-compose

COPY --from=builder /out/trest-installer /usr/local/bin/trest-installer
COPY config.yaml /etc/trest/config.yaml

WORKDIR /work
EXPOSE 9091

ENTRYPOINT ["/usr/local/bin/trest-installer"]
CMD ["audit", "--config", "/etc/trest/config.yaml", "--repo-dir", "/work/repo"]
