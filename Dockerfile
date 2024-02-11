FROM golang:1.21
WORKDIR /app
COPY go.mod go.sum ./
LABEL authors="splatjov"

RUN go mod download


COPY *.go ./
COPY *.db ./
COPY *.session ./

RUN CGO_ENABLED=1 GOOS=linux go build -o /valentinkaBotBackend

EXPOSE 3000

CMD ["/valentinkaBotBackend"]