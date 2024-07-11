# # Stage 1: Build the application
FROM golang:1.22-alpine as builder

RUN apk add --no-cache upx

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /libp2p-node cmd/oracle/main.go
# Compress the binary
RUN upx --best --lzma /libp2p-node

# # Stage 2: Create the final minimal image
FROM gcr.io/distroless/static
# Create a non-root user and group
USER nonroot:nonroot
COPY --from=builder /libp2p-node /libp2p-node
COPY --from=builder /app/config/config.yaml .

# Use a non-root user
USER nonroot:nonroot

CMD [ "/libp2p-node" ]
