.PHONY: test vet list tone capture-play sink-play build run install

test:
	go test -race ./...

vet:
	go vet ./...

build:
	mkdir -p bin
	go build -o bin/multiout ./cmd/multiout

run:
	go run ./cmd/multiout

install: build
	install -d $(HOME)/.local/bin
	install -m 755 bin/multiout $(HOME)/.local/bin/multiout
	install -d $(HOME)/.local/share/applications
	install -m 644 multiout.desktop $(HOME)/.local/share/applications/multiout.desktop
	install -d $(HOME)/.local/share/icons
	install -m 644 internal/ui/multiout-logo-rounded.png $(HOME)/.local/share/icons/multiout.png
	install -d $(HOME)/.local/share/icons/hicolor/512x512/apps
	install -m 644 internal/ui/multiout-logo-rounded.png $(HOME)/.local/share/icons/hicolor/512x512/apps/multiout.png
	@if command -v gtk-update-icon-cache >/dev/null 2>&1; then \
		gtk-update-icon-cache -f -t $(HOME)/.local/share/icons/hicolor 2>/dev/null || true; \
	fi
	@if command -v update-desktop-database >/dev/null 2>&1; then \
		update-desktop-database $(HOME)/.local/share/applications 2>/dev/null || true; \
	fi
	@echo "Installed multiout to $(HOME)/.local/bin and desktop launcher created with rounded logo!"

list:
	go run ./cmd/spike list

# usage: make tone A=0 B=1
tone:
	go run ./cmd/spike tone $(A) $(B)

# usage: make capture-play C=1 P=1
capture-play:
	go run ./cmd/spike capture-play $(C) $(P)

# usage: make sink-play DEVS="1 2"
sink-play:
	go run ./cmd/spike sink-play $(DEVS)
