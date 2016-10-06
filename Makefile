VERSION=$(shell perl -wnE 'say $$1 if /Version\s=\s"(.*?)\"/' main.go)

deck:
	go build -ldflags "-linkmode external -extldflags -static"
dist: deck
	strip deck
	mv deck deck-v$(VERSION)-$$(uname | tr [A-Z] [a-z])-$$(uname -m)-static
	xz deck-v$(VERSION)-$$(uname | tr [A-Z] [a-z])-$$(uname -m)-static
clean:
	rm -f deck
	rm -f deck-v*
