FROM       alpine:3.18
RUN        apk add --update --no-cache ca-certificates
COPY       auth-gateway /usr/bin/auth-gateway
EXPOSE     80
ENTRYPOINT [ "/usr/bin/auth-gateway" ]
