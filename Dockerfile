FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./run-migrations cmd/database/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o ./srs cmd/api/main.go

EXPOSE 8080

CMD ./run-migrations && ./srs