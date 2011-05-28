include $(GOROOT)/src/Make.inc

.PHONY: all install clean

# ugly ugly ugly

all:
	gomake -C api all
	gomake -C api install

	gomake -C mods all
	gomake -C mods install

	gomake -C rocket all

	gomake -C bot all

install:
	gomake -C api all
	gomake -C api install

	gomake -C mods all
	gomake -C mods install

	gomake -C rocket all
	gomake -C rocket install

	gomake -C bot all
	gomake -C bot install

clean:
	gomake -C api clean
	gomake -C mods clean
	gomake -C rocket clean
	gomake -C bot clean

fmt:
	gofmt -w=1 .
