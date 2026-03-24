FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .


FROM alpine:3.22
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/app .
COPY config.yaml .
COPY rate_rules.json .
COPY scripts/ ./scripts/
COPY _proto/ ./_proto/
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080
CMD ["./app"]
