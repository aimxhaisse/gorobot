all:
	make -f MakeBot $@
	make -f MakeMods $@

clean:
	make -f MakeBot $@
	make -f MakeMods $@
