FROM ubuntu:14.04
MAINTAINER s. rannou <mxs@sbrk.org>, Manfred Touron <m@42.am>

ENV DEBIAN_FRONTEND noninteractive
ENV USERNAME gorobot

# deps
RUN apt-get update && \
    apt-get upgrade -q -y && \
    apt-get install -q -y \
    git netcat golang bc build-essential

# build
RUN mkdir /usr/src/app
ADD . /usr/src/app
WORKDIR /usr/src/app
RUN go build

# admin port for commands
EXPOSE 3112

# here we go
CMD ./gorobot -c root/gorobot.json
