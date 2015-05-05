FROM ubuntu:14.04
MAINTAINER s. rannou <mxs@sbrk.org>, Manfred Touron <m@42.am>

ENV DEBIAN_FRONTEND noninteractive
ENV USERNAME gorobot

# deps
RUN apt-get update && \
    apt-get upgrade -q -y && \
    apt-get install -q -y \
    git netcat golang bc build-essential

# user
RUN groupadd $USERNAME -g 1013 && \
    useradd -m $USERNAME -u 1013 -g 1013

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
CMD ./start.sh
