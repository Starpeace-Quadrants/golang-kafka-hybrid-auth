FROM golang:1.19 as builder

WORKDIR /opt/app/api

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./

FROM golang:1.19 as final
WORKDIR /opt/app/api

COPY --from=builder /opt/app/api/auth-service /usr/local/bin/auth-service

ENTRYPOINT ["/usr/local/bin/auth-service"]