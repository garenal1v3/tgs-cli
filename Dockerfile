FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
      -X github.com/searchtgcli/tgs/internal/cmd.Version=${VERSION} \
      -X github.com/searchtgcli/tgs/internal/cmd.Commit=${COMMIT} \
      -X github.com/searchtgcli/tgs/internal/cmd.Date=${DATE}" \
    -o /tgs ./cmd/tgs/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /tgs /usr/local/bin/tgs
ENTRYPOINT ["tgs"]
