FROM golang:1.12-alpine AS mods
RUN apk add --no-cache git
ADD go.mod /usr/src/rancher-memory-limiter/go.mod
ADD go.sum /usr/src/rancher-memory-limiter/go.sum
WORKDIR /usr/src/rancher-memory-limiter
RUN go mod download

FROM golang:1.12-alpine AS build
ADD . /usr/src/rancher-memory-limiter/
COPY --from=mods /go/pkg/mod/ /go/pkg/mod/
WORKDIR /usr/src/rancher-memory-limiter

RUN go install rancher-memory-limiter


FROM alpine
COPY --from=build /go/bin/rancher-memory-limiter /usr/bin/rancher-memory-limiter

CMD ["/usr/bin/rancher-memory-limiter"]
