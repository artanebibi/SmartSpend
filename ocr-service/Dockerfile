# Build
FROM golang:1.24-bullseye AS builder
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    libtesseract-dev \
    libleptonica-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build statically linked binary
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./api

# Run
FROM debian:bullseye-slim
# runtime tesseract
RUN apt-get update && apt-get install -y tesseract-ocr && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/main .
COPY mkd.traineddata /usr/share/tesseract-ocr/4.00/tessdata/
CMD ["./main"]
