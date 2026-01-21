.PHONY: build install clean

build:
	go build -o economist .
	@if [ "$$(uname)" = "Darwin" ]; then codesign -f -s - economist 2>/dev/null || true; fi

install: build
	cp economist ~/bin/
	@if [ "$$(uname)" = "Darwin" ]; then codesign -f -s - ~/bin/economist 2>/dev/null || true; fi

clean:
	rm -f economist
