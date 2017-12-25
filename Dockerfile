FROM gcr.io/google-containers/debian-base-amd64:0.3

RUN clean-install \
  bash \
  curl \
  socat \
  haproxy \
  ca-certificates \
  liblua5.3-0 \
  dumb-init

RUN mkdir -p /etc/haproxy/errors /var/state/haproxy
RUN for ERROR_CODE in 400 403 404 408 500 502 503 504;do curl -sSL -o /etc/haproxy/errors/$ERROR_CODE.http \
	https://raw.githubusercontent.com/haproxy/haproxy-1.5/master/examples/errorfiles/$ERROR_CODE.http;done

ADD haproxy-init /
ADD haproxy_reload /

RUN touch /var/run/haproxy.pid

ENTRYPOINT ["/usr/bin/dumb-init"]

CMD ["/haproxy-init"]
