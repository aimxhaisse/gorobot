FROM ubuntu:latest
MAINTAINER s. rannou <mxs@sbrk.org>, Manfred Touron <m@42.am>

ENV DEBIAN_FRONTEND noninteractive

# deps
RUN echo "deb http://archive.ubuntu.com/ubuntu precise main universe" > /etc/apt/sources.list && \
    apt-get update && \
    apt-get upgrade && \
    apt-get install -qq -y \
    golang git binutils gcc netcat bc

# user
RUN groupadd gorobot && \
    useradd -m gorobot -g gorobot
USER gorobot

# build
ADD . /home/gorobot/gorobot/
WORKDIR /home/gorobot/gorobot/

RUN sed -i s/mxs/$USER/ all.bash && \
    ./all.bash build

# admin port for commands
EXPOSE 3112

# here we go
CMD ./all.bash start && tail -F ./root/gorobot.log
