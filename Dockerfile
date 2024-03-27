FROM golang:1.21.6-alpine3.19

WORKDIR /build

#RUN apk --no-cache add build-base
RUN apk add build-base

ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
#RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o looklapi .
RUN go build -a -o looklapi .

FROM alpine:3.19.1

#RUN apk --no-cache add tzdata ca-certificates libc6-compat libgcc libstdc++
RUN apk add tzdata ca-certificates libc6-compat libgcc libstdc++

WORKDIR /app
COPY --from=0 /build/looklapi /app/
COPY --from=0 /build/application.yml /app/
COPY --from=0 /build/application-dev.yml /app/
COPY --from=0 /build/application-prod.yml /app/

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone

EXPOSE 8001
ENTRYPOINT ["/app/looklapi"]