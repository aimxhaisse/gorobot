include $(GOROOT)/src/Make.inc

.PHONY: all install clean

all:
	gomake -C api all
	gomake -C mods all
	gomake -C rocket all
	gomake -C bot all

install: all
	gomake -C api install
	gomake -C mods install
	gomake -C rocket install
	gomake -C bot install

clean:
	gomake -C api clean
	gomake -C mods clean
	gomake -C rocket clean
	gomake -C bot clean

fmt:
	gofmt -w=1 .
