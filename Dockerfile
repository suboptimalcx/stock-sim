# build stage
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
SHELL ["/bin/sh", "-e", "-c"]
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH

# statically linked binary 
# maybe add -ldflags="-w -s"
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o main ./cmd/api/main.go 

# run stage
FROM alpine:3.22.4
WORKDIR /app

RUN adduser -D appuser

COPY --from=builder /app/main .

USER appuser

EXPOSE 8080

CMD ["./main"]