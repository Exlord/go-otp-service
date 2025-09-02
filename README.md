# OTP Service (Go)

A minimal backend service implementing OTP-based login/registration with JWT, rate limiting,
and basic user management. Implements in-memory storage for simplicity (no external DB).

## Why in-memory?
- Fast to set up and easy to understand.
- Fits requirements: temporary OTP storage and simple user data.
- Trade-off: Data resets on restart. Swap in a real DB later by implementing the repository interfaces.

## Run locally
```bash
# prerequisites: Go 1.22+
export JWT_SECRET=devsecret
go run ./cmd/server
```
Visit: http://localhost:8080/health and http://localhost:8080/openapi for API docs (serves openapi.yaml).

## Run with Docker
```bash
docker build -t otpservice:latest .
docker run --rm -p 8080:8080 -e JWT_SECRET=devsecret otpservice:latest
```

## Example flow (curl)
```bash
# 1) Request OTP (printed to server console)
curl -X POST http://localhost:8080/api/v1/auth/request-otp   -H "Content-Type: application/json"   -d '{"phone":"+905551112233"}'

# 2) Verify OTP (replace 123456 with the code from console)
curl -X POST http://localhost:8080/api/v1/auth/verify   -H "Content-Type: application/json"   -d '{"phone":"+905551112233","otp":"123456"}'

# => {"token":"<JWT>","user":{"id":"...","phone":"+905...","registeredAt":"..."}}}

# 3) Authenticated user endpoints
TOKEN=<JWT>
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/users
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/users/<id>
```

## Rate limiting
- Max **3 OTP requests** per phone within **10 minutes**. Exceeding returns HTTP 429.

## OTP rules
- 6-digit numeric, expires after 2 minutes, stored in-memory.

## API Docs
- OpenAPI served at `/openapi` and raw YAML at `/openapi.yaml`.
