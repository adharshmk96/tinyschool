FROM golang:1.26-bookworm AS build

WORKDIR /src
COPY tinyschool-api/go.mod tinyschool-api/go.sum ./
RUN go mod download
COPY tinyschool-api/ ./
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/tinyschool-api .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir /data \
    && chown 65532:65532 /data
COPY --from=build /out/tinyschool-api /usr/local/bin/tinyschool-api

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["tinyschool-api"]
