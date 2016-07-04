FROM scratch
ADD bin/serve-http /serve-http
CMD ["/serve-http"]
