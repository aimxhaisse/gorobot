include $(GOROOT)/src/Make.inc

.PHONY: all install clean

all:
	gomake -C broadcast
	gomake -C example
	gomake -C radio
	gomake -C rss
	gomake -C scripts
	gomake -C userjoin

install: all
	gomake -C broadcast
	gomake -C example
	gomake -C radio
	gomake -C rss
	gomake -C scripts
	gomake -C userjoin

clean:
	gomake -C broadcast
	gomake -C example
	gomake -C radio
	gomake -C rss
	gomake -C scripts
	gomake -C userjoin
