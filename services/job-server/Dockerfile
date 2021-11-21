FROM golang as builder

WORKDIR /app/go/src/github.com/bogdan-copocean/hasty-server

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main services/job-server/main.go

FROM alpine:3.14

COPY --from=builder /app/go/src/github.com/bogdan-copocean/hasty-server/main .

CMD [ "./main" ]