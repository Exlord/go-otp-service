# ---------- build ----------
FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/otpservice ./cmd/server

# ---------- run ----------
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /bin/otpservice /otpservice
COPY openapi.yaml /openapi.yaml
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/otpservice"]
