FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

COPY ./frontend/package.json ./

RUN npm install --omit=dev

COPY ./frontend .

RUN npm run build

# Backend
FROM golang:1.24-alpine AS backend-builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o ./run-migrations cmd/database/main.go

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o ./srs cmd/api/main.go

FROM alpine:3.19

COPY --from=frontend-builder /frontend/dist /frontend/dist

WORKDIR /app

RUN mkdir -p /app/database/migrations

COPY --from=backend-builder /build/run-migrations /build/srs ./

COPY --from=backend-builder /build/database/migrations/*.sql /app/database/migrations/

RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["sh", "-c", "./run-migrations && ./srs"]
