FROM golang:1.22 AS build-env

WORKDIR /go/src/app
COPY . /go/src/app

RUN go build -ldflags "-s -w" -o /go/bin/app main.go

FROM gcr.io/distroless/base:nonroot
COPY --from=build-env --chown=nonroot:nonroot /go/bin/app /

USER nonroot

CMD ["/app"]
