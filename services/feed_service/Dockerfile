FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o app .


FROM alpine:3.22
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY --from=builder /app/app .
RUN chown -R appuser:appgroup /app                                           
USER appuser 

EXPOSE 50051
CMD ["./app"]
