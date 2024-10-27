FROM golang:1.23.2-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build cmd/api/main.go

FROM builder
WORKDIR /app
COPY --from=builder /app/main .
COPY .env .
COPY credentials.json .

EXPOSE 8080

CMD ["./main"]