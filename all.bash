#!/usr/bin/env bash
#
# author: s. rannou <mxs@sbrk.org>
# contributor: Manfred Touron <m@42.am>

PROJECT="gorobot"

# user to run as
USER="gorobot"

# edit these to your needs if you know what you are doing
CHDIR="root"			# directory where we chdir/chroot
PID="$PROJECT.pid"		# relative to $CHDIR
BIN="$PROJECT"			# relative to $CHDIR
LOG="gorobot.log"		# relative to $CHDIR
CONFIG="gorobot.json"		# relative to $CHDIR
LOGS="logs"			# relative to $CHDIR

cd $(dirname $0)

function ok {			# <msg>
    echo -e "\033[0;32;49m$@\\033[0m"
    return 0
}

function ko {			# <msg>
    echo -e "\033[0;31;49mError: $@\\033[0m"
    return 1
}

function title {		# <msg>
    echo -e "\033[0;32;40m$@\\033[0m"
    return 0
}

function warn-upon-failure {	# <return_value> <msg>
    local ret=$1
    local msg=$2
    if [ $ret -ne 0 ]
    then
	ko "$msg"
    fi
    return $ret
}

function be-quiet {		# <command1>
    $@ 2>&1 > /dev/null
    return $?
}


function get-uid {
    return $(id -u)
}

function usage {
    echo -e "usage: $0 [build|stop|start|restart|status]"
}

function check-env {
    # make sure $DEPLOY_DIR exists
    if ! [ -d $DEPLOY ]
    then
	warn-upon-failure 1 "$DEPLOY does not exist, maybe you should edit \$DEPLOY" || return 1
    fi

    # make sure $USER exists
    be-quiet id $USER
    warn-upon-failure $? "User $USER does not exist, maybe you should edit \$USER"

    return 0
}

check-env || exit 1

case $1 in
    "stop")
	ok "stopping $PROJECT"
	if [ -f $CHDIR/$PID ]
	then
	    kill -9 $(cat $CHDIR/$PID)
	    rm -f $CHDIR/$PID
	fi
	ok "daemon stopped"
	exit 0
	;;

    "start")
	ok "starting $PROJECT"
	mkdir -p $CHDIR/$LOGS
	./daemonize -p $PID -u $USER -c $CHDIR -- ./$PROJECT -c $CONFIG
	warn-upon-failure $? "Can't start the daemon, check your config" || exit 1
	ok "daemon started"
	exit 0
	;;

    "restart")
	$0 stop
	$0 start
	exit 0
	;;

    "status")
        pid=$(ps -o pid= --pid $(cat $CHDIR/$PID 2>/dev/null) 2>/dev/null)
        if [ "$pid" != "" ]
        then
            ok "$PROJECT is running, with pid $pid"
        else
            ko "$PROJECT is down"
        fi
        exit 0
        ;;

    "log")
	if [ -f $CHDIR/$LOG ]
	then
	    ok "last 25 log entries:"
	    tail -n 25 $CHDIR/$LOG
	else
	    ko "can't find log file ($CHDIR/$LOG)"
	fi
	exit 0
	;;

    "build")
	# build the daemonizer
	if [ ! -f daemonize ] || [ $(stat -c %Y daemonize) -lt $(stat -c %Y daemonizer/daemonizer.c) ]
	then
	    ok "building daemonize..."
	    gcc -W -Wall -pedantic -ansi -O3 -Wno-unused-result daemonizer/daemonizer.c -o daemonize
	    warn-upon-failure $? "can't build daemonize"
	    ok "daemonize built (or not)"
	else
	    ok "daemonize is already up-to-date (skipped)"
	fi
	
	# build the website
	ok "building $PROJECT..."
	go build -o $CHDIR/$PROJECT
	warn-upon-failure $? "unable to build $PROJECT"
	ok "$PROJECT built (or not)..."
	exit 0
	;;
esac

usage
exit 1
