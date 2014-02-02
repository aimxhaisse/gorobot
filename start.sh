#!/bin/bash

cp gorobot root/gorobot && \
    touch ./root/gorobot.log && \
    chown -R $USERNAME . && \
    ./all.bash start && \
    tail -F ./root/gorobot.log
