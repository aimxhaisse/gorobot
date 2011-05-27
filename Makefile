include $(GOROOT)/src/Make.inc

.PHONY: all install clean

all:
	gomake -C api
	gomake -C bot
	gomake -C mods
	gomake -C rocket

install: all
	gomake -C api
	gomake -C bot
	gomake -C mods
	gomake -C rocket

clean:
	gomake -C api
	gomake -C bot
	gomake -C mods
	gomake -C rocket
