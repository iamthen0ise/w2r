FROM golang:1.17-alpine AS builder

ENV GO111MODULE=on

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -a -o main .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /root/

COPY --from=builder /build/main .

ENTRYPOINT ["/root/main"]

