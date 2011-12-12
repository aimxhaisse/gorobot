include $(GOROOT)/src/Make.inc

.PHONY: all install clean

all:
	gomake -C broadcast all
	gomake -C example all
	gomake -C radio all
#	gomake -C rss all
	gomake -C scripts all
	gomake -C userjoin all

install: all
	gomake -C broadcast install
	gomake -C example install
	gomake -C radio install
#	gomake -C rss install
	gomake -C scripts install
	gomake -C userjoin install

clean:
	gomake -C broadcast clean
	gomake -C example clean
	gomake -C radio clean
#	gomake -C rss clean
	gomake -C scripts clean
	gomake -C userjoin clean
