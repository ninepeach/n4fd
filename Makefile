BINARY := n4fd

.PHONY: build run clean

build:
	go build -o $(BINARY) ./cmd/proxy

run: build
	./$(BINARY) \
	  -listen :8443 \
	  -priv 'DBk2n9pzy8e2CwsXU0WrZ3Zobldz_467c2eI-Z0iXw0' \
	  -dest 'update.microsoft.com:443' \
	  -sni 'update.microsoft.com' \
	  -short 'a1b2c3d4e5f67890' \
	  -uuids 0d32d1f2-3453-429f-82cf-8cabcca1223e \
	  -debug

clean:
	rm -f $(BINARY)
