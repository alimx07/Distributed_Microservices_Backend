FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o app .


FROM alpine:3.22
RUN apk add --no-cache libc6-compat
WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE 50051
CMD ["./app"]
