FROM golang:1.18-alpine as builder

COPY . /eventops

WORKDIR /eventops

ENV CGO_ENABLED 0

RUN export GO111MODULE=on && export GOPROXY=https://goproxy.cn && GOOS=linux && GOARCH=amd64 && go build -a -ldflags '-extldflags "-static"' -o /eventops/bin/eventops /eventops/cmd/.

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.6

COPY --from=builder /eventops/bin/eventops /bin/eventops

RUN chmod +x /bin/eventops

CMD eventops