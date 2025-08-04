.PHONY: server
server:
	go run ./cmd/server

.PHONY: wasm
wasm:
	GOOS=js GOARCH=wasm go build  -o ./assets/main.wasm ./src/core/wasm.go

.PHONY: tinywasm
tinywasm:
	GOOS=js GOARCH=wasm tinygo build -o ./assets/wasm.wasm ./src/core/wasm.go

.PHONY: shell
shell:
	nix develop -c $$SHELL


