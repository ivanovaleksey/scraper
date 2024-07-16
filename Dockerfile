FROM golang:1.22-alpine AS builder

RUN apk add --update make

WORKDIR /build

COPY go.* .
RUN go mod download

COPY Makefile Makefile

COPY cmd/ cmd/
COPY internal/ internal/
RUN make build

FROM alpine

WORKDIR /app

COPY --from=builder /build/bin/scraper .

ENTRYPOINT ["/app/scraper"]
CMD ["-web"]
