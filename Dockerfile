FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy denpendency
COPY go.mod go.sum ./
RUN go mod download

# Copy Code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# -------------------------------

FROM gcr.io/distroless/static-debian12

WORKDIR /root/

# Copy Executable
COPY --from=builder /app/main .
# Copy .env
COPY --from=builder /app/.env .

# Port
EXPOSE 8080

# Run
CMD [ "./main" ]
