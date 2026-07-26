# Stage 1: Build
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git make
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/ankurCES/floc-zure/internal/cli.Version=$(git describe --tags --always --dirty 2>/dev/null || echo docker)" -o /bin/azfloci ./cmd/azfloci
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /bin/az-simulator ./simulator/cmd/az

# Stage 2: Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/azfloci /usr/local/bin/azfloci
COPY --from=builder /bin/az-simulator /usr/local/bin/az
# Pre-configure simulator as the az backend
ENV AZFLOCI_AZ_PATH=/usr/local/bin/az
ENV AZFLOCI_SIM_STATE=/data/state.json
RUN mkdir -p /data /workspace
WORKDIR /workspace
ENTRYPOINT ["azfloci"]
CMD ["--help"]
