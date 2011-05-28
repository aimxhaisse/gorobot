include $(GOROOT)/src/Make.inc

.PHONY: all install clean

all:
	gomake -C api $@
	gomake -C mods $@
	gomake -C rocket $@
	gomake -C bot $@

install: all
	gomake -C api $@
	gomake -C mods $@
	gomake -C rocket $@
	gomake -C bot $@

clean:
	gomake -C api $@
	gomake -C mods $@
	gomake -C rocket $@
	gomake -C bot $@

fmt:
	gofmt -w=1 .
