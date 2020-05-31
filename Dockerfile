FROM golang:1.14-alpine as build-golang

ENV GOPROXY="https://goproxy.cn/,https://mirrors.aliyun.com/goproxy/,https://goproxy.baidu.com/,direct"
ENV GOSUMDB=off

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app cli/main.go

FROM gruebel/upx:latest as upx
COPY --from=build-golang /app /app.org

RUN upx --best --lzma -o /app /app.org

FROM alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache tzdata

COPY --from=upx /app /app

EXPOSE 443

ENTRYPOINT ["/app"]
