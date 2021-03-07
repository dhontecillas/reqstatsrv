FROM golang:1.16.0-buster as builder

COPY . /go/src/github.com/dhontecillas/reqstatsrv
WORKDIR /go/src/github.com/dhontecillas/reqstatsrv
RUN go build ./cmd/reqstatsrv

# FROM scratch
FROM bitnami/minideb:stretch

EXPOSE 9876
COPY --from=builder /go/src/github.com/dhontecillas/reqstatsrv/reqstatsrv /reqstatsrv

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.description="ReqStatSrv: Rate Limiting Proxy" \
      org.label-schema.name="reqstatsrv" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="http://www.hontecillas.com" \
      org.label-schema.vcs-url="https://github.com/dhontecillas/dynlimits" \
      org.label-schema.vcs-ref=$BUILD_VCS_REF \
      org.label-schema.vendor="David Hontecillas" \
      org.label-schema.version=$BUILD_VERSION

CMD ["/reqstatsrv"]
