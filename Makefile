.PHONY: build install clean

build:
	go build -o economist .
	@if [ "$$(uname)" = "Darwin" ]; then codesign -s - economist; fi

install: build
	cp economist ~/bin/

clean:
	rm -f economist
