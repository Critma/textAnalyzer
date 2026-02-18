# Build
FROM golang:1.26.0-alpine AS builder
WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./main ./cmd/*.go

#Run
FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]