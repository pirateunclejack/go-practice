FROM golang:alpine as builder

RUN mkdir /build
ADD . /build

WORKDIR /build

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn

RUN go build -o main

# stage 2
FROM alpine

RUN adduser -S -D -h /app appuser

USER appuser

COPY . /app

COPY --from=builder /build/main /app/

WORKDIR /app/

EXPOSE 3000

CMD [ "/app/main" ]
