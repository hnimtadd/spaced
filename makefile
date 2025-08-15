.PHONY: server
server:
	@ go run ./cmd/server ui

.PHONY: wasm
wasm:
	@GOOS=js GOARCH=wasm go build  -o ./ui/assets/session.wasm ./wasm/session/session.go
	@GOOS=js GOARCH=wasm go build  -o ./ui/assets/stats.wasm ./wasm/stats/stats.go

.PHONY: shell
shell:
	nix develop -c $$SHELL


