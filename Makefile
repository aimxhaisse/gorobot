include $(GOROOT)/src/Make.inc

DEPS = gorobot scripts rss example broadcast
TARG = bin/m1ch3l
GOFILES = main.go

include $(GOROOT)/src/Make.cmd
