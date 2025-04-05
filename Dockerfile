# base image
FROM golang:1.23.4-alpine as base
WORKDIR /emailer

ENV CGO_ENABLED=0
COPY go.mod go.sum /emailer/
RUN go mod download

ADD . .
RUN go build -o /usr/local/bin/emailer ./cmd/emailer

# runner image
FROM gcr.io/distroless/static:latest
WORKDIR /app
COPY --from=base /usr/local/bin/emailer emailer
ENTRYPOINT ["/app/emailer"]
