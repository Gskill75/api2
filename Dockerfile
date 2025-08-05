FROM golang:alpine AS ca-certificates
FROM scratch
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./api /usr/bin/api

ENTRYPOINT [ "/usr/bin/api" ]
CMD []
