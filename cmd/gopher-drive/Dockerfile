FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/gopher-drive ./cmd/gopher-drive/main.go

FROM alpine:3.23.4 AS runner
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /app/gopher-drive .
ENV PROJECT_ROOT=/app
CMD ["./gopher-drive"]