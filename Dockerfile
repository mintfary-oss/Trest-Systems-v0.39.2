# Default CLI image; libc/libpq must match the Linux prebuilt binary.
FROM debian:bookworm-slim
ARG TARGETARCH=amd64
RUN apt-get -o Acquire::Retries=3 -o Acquire::http::Timeout=30 update && apt-get install -y --no-install-recommends ca-certificates tzdata libpq5 && rm -rf /var/lib/apt/lists/* && useradd -r -u 10001 trest
WORKDIR /app
COPY release/bin/linux/${TARGETARCH}/trest /usr/local/bin/trest
COPY proektirovka-sdaniy/configs /app/proektirovka-sdaniy/configs
RUN chmod 0755 /usr/local/bin/trest && chown -R trest:trest /app
USER trest
ENTRYPOINT ["/usr/local/bin/trest"]
