.PHONY: test vet list tone capture-play sink-play build run install deb exe

VERSION ?= 1.0.0
DEB_NAME = multiout_$(VERSION)_amd64

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

deb: build
	rm -rf build/deb/$(DEB_NAME)
	mkdir -p build/deb/$(DEB_NAME)/DEBIAN
	mkdir -p build/deb/$(DEB_NAME)/usr/bin
	mkdir -p build/deb/$(DEB_NAME)/usr/share/applications
	mkdir -p build/deb/$(DEB_NAME)/usr/share/icons/hicolor/512x512/apps
	mkdir -p build/deb/$(DEB_NAME)/usr/share/icons
	install -m 755 bin/multiout build/deb/$(DEB_NAME)/usr/bin/multiout
	install -m 644 multiout.desktop build/deb/$(DEB_NAME)/usr/share/applications/multiout.desktop
	install -m 644 internal/ui/multiout-logo-rounded.png build/deb/$(DEB_NAME)/usr/share/icons/multiout.png
	install -m 644 internal/ui/multiout-logo-rounded.png build/deb/$(DEB_NAME)/usr/share/icons/hicolor/512x512/apps/multiout.png
	@echo "Package: multiout" > build/deb/$(DEB_NAME)/DEBIAN/control
	@echo "Version: $(VERSION)" >> build/deb/$(DEB_NAME)/DEBIAN/control
	@echo "Section: sound" >> build/deb/$(DEB_NAME)/DEBIAN/control
	@echo "Priority: optional" >> build/deb/$(DEB_NAME)/DEBIAN/control
	@echo "Architecture: amd64" >> build/deb/$(DEB_NAME)/DEBIAN/control
	@echo "Maintainer: Sagar <codesbysagar>" >> build/deb/$(DEB_NAME)/DEBIAN/control
	@echo "Depends: pipewire (>= 0.3), wireplumber" >> build/deb/$(DEB_NAME)/DEBIAN/control
	@echo "Description: Simultaneous Multi-Device Audio Output" >> build/deb/$(DEB_NAME)/DEBIAN/control
	@echo " Play audio to multiple earphones simultaneously with drift compensation." >> build/deb/$(DEB_NAME)/DEBIAN/control
	dpkg-deb --build --root-owner-group build/deb/$(DEB_NAME) bin/$(DEB_NAME).deb
	@echo "Created Debian package: bin/$(DEB_NAME).deb"

exe:
	@if ! command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then \
		echo "Error: x86_64-w64-mingw32-gcc not found. Please run: sudo apt install -y gcc-mingw-w64-x86-64"; \
		exit 1; \
	fi
	mkdir -p bin
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags="-H=windowsgui -s -w" -o bin/multiout.exe ./cmd/multiout
	@echo "Created Windows executable: bin/multiout.exe"

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
