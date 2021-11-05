FROM golang:1.17 as build

ENV BIN_FILE /opt/minipic
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

ARG LDFLAGS
RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o ${BIN_FILE} cmd/*

FROM alpine:3.14.2

LABEL SERVICE="minipic"
LABEL MAINTAINERS="sumin.200ok@gmail.com"


ENV BIN_FILE "/opt/minipic"
COPY --from=build ${BIN_FILE} ${BIN_FILE}
COPY test /opt/test

ENV CONFIG_FILE /etc/minipic/config.toml
COPY ./configs/config.toml ${CONFIG_FILE}
WORKDIR /opt/
CMD ${BIN_FILE} -config ${CONFIG_FILE}
