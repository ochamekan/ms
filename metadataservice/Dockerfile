FROM golang:1.25

WORKDIR /usr/src/app

COPY go.mod go.sum .env ./

RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/app ./metadataservice/cmd/main.go

ENTRYPOINT app
