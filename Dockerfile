FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /sentinel .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /sentinel /usr/local/bin/sentinel
ENTRYPOINT ["sentinel"]
