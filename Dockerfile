FROM golang:1.12-alpine3.9 AS builder

WORKDIR /app
RUN apk add --no-cache --virtual .go-deps git gcc musl-dev openssl pkgconf

COPY go.mod .
COPY go.sum .
COPY mibs/ mibs/
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o traplogger -ldflags '-w -s' main.go

# Build the smallest image possible
FROM scratch AS runner
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/mibs/ mibs/
COPY --from=builder /app/traplogger /traplogger

ENTRYPOINT [ "/traplogger" ]
