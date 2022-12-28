
.PHONY: build

build:
	wails build

install: build
	mkdir -p ~/.local/share/applications ~/.local/share/icons/hicolor/256x256/apps ~/.local/bin/
	cp build/bin/tailscale-client ~/.local/bin
	cp tailscale-client.desktop ~/.local/share/applications/
	cp icon/on.png ~/.local/share/icons/hicolor/256x256/apps/com.tailscale.png
