FROM golang:1.24

WORKDIR ./app

COPY . .

RUN go mod tidy && go build -o scoreboard ./cmd/scoreboard/main.go

EXPOSE 8080

CMD ["./scoreboard"]