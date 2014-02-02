Gorobot
=======

Irc robot in GO

Docker
------

    # Build
    docker build -t aimxhaisse/gorobot .

    # Run in foreground
    docker run -i -t -rm aimxhaisse/gorobot

    # Run in background
    docker run -d aimxhaisse/gorobot

    # Mounts scripts directory for dev
    docker run -i -t -rm \
    	   -v $(pwd)/root/ /home/gorobot/gorobot/root/ \
    	   aimxhaisse/gorobot
    	   
Extending with Docker
---------------------

    FROM aimxhaisse/gorobot
    ADD . ./root
    ...
