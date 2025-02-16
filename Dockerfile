FROM golang:1.23.5 AS builder

WORKDIR /usr/src/app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/ .


FROM scratch
COPY --from=builder /usr/local/bin/chat /
CMD ["/chat"]