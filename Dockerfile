FROM golang:1.26-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/local/bin/deploy ./deploy && \
    go build -o /usr/local/bin/delete ./delete && \
    go build -o /usr/local/bin/archive ./archive && \
    go build -o /usr/local/bin/unarchive ./unarchive
