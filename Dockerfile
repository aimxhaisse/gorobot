FROM ubuntu:latest
MAINTAINER s. rannou <mxs@sbrk.org>, Manfred Touron <m@42.am>

ENV DEBIAN_FRONTEND noninteractive
ENV USERNAME gorobot

# deps
RUN echo "deb http://archive.ubuntu.com/ubuntu precise main universe" > /etc/apt/sources.list && \
    apt-get update && \
    apt-get upgrade && \
    apt-get install -qq -y \
    golang git binutils gcc netcat bc

# user
RUN groupadd $USERNAME && \
    useradd -m $USERNAME -g $USERNAME

# build
ADD . /tmp/gorobot/
RUN mv /tmp/gorobot /home/$USERNAME/gorobot && \
    mkdir -p /home/$USERNAME/gorobot/root/logs && \
    touch /home/$USERNAME/gorobot/root/gorobot.log && \
    chown -R $USERNAME /home/$USERNAME/gorobot
# $USERNME not expanded in WORKDIR
WORKDIR /home/gorobot/gorobot/

RUN sed -i s/mxs/$USERNAME/ ./all.bash && \
    ./all.bash build

# admin port for commands
EXPOSE 3112

# here we go
CMD touch ./root/gorobot.log && \
    chown -R $USERNAME . && \
    ./all.bash start && \
    tail -F ./root/gorobot.log
