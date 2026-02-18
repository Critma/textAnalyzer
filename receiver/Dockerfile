# Build
FROM golang:1.26.0-alpine AS builder
WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download
COPY . .

# install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest
# gen swagger
RUN swag init -g ./cmd/main.go -o cmd/docs

# compile
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./main ./cmd/*.go

#Run
FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/main .
# COPY --from=builder /app/cmd/docs ./docs
CMD ["./main"]