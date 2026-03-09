.PHONY: build test wasm clean

build:
	go build -o oh-my-line ./cmd/oh-my-line/

test:
	go test ./...

wasm:
	GOOS=js GOARCH=wasm go build -o renderer.wasm ./cmd/wasm/

clean:
	rm -f oh-my-line renderer.wasm
