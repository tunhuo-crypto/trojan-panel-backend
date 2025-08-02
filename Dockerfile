FROM alpine:3.15
LABEL maintainer="jonsosnyan <https://jonssonyan.com>"
WORKDIR /tpdata/trojan-panel/
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
COPY build/trojan-panel-${TARGETOS}-${TARGETARCH}${TARGETVARIANT} trojan-panel
# Set apk China mirror
# RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add bash tzdata ca-certificates && \
    rm -rf /var/cache/apk/*
ENTRYPOINT chmod 777 ./trojan-panel && \
    ./trojan-panel \
    -host=${mariadb_ip} \
    -port=${mariadb_port} \
    -user=${mariadb_user} \
    -password=${mariadb_pas} \
    -redisHost=${redis_host} \
    -redisPort=${redis_port} \
    -redisPassword=${redis_pass} \
    -serverPort=${server_port}
