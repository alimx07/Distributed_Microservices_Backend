FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .


FROM alpine:3.22
WORKDIR /root/
COPY --from=builder /app/app .


COPY config.yaml .
COPY rate_rules.json .
COPY ./scripts ./

COPY _proto/ ./_proto/

EXPOSE 8080
CMD ["./app"]
