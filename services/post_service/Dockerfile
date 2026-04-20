FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .


FROM alpine:3.22
RUN addgroup -S appgroup && adduser -S appuser -G appgroup 
WORKDIR /root/
COPY --from=builder /app/app .
RUN chown -R appuser:appgroup /app                                           
USER appuser 

EXPOSE 50051
CMD ["./app"]
