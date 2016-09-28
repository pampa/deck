default:
	go build -ldflags "-linkmode external -extldflags -static"
	install -m 755 deck /usr/bin/deck
clean:
	rm -f deck
