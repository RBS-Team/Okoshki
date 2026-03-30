FROM golang:1.25-alpine AS builder

RUN apk add --no-cache tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN swag init -g cmd/core/main.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/core ./cmd/core/main.go

# --- Runner ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /
COPY --from=builder /bin/core /core
EXPOSE 8080
CMD ["/core", "-f=/config"]