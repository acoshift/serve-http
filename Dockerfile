FROM scratch
ADD bin/serve-http /serve-http
ENTRYPOINT ["/serve-http"]
