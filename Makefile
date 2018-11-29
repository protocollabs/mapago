TARGET := mapago
PACKAGE := github.com/protocollabs/mapago
DATE    := $(shell date +%FT%T%z)
VERSION := $(shell git describe --tags --always --dirty)
GOBIN   :=$(GOPATH)/bin

MKDIR_P = @mkdir -p

LDFLAGS  = "-X $(PACKAGE)/core.BuildVersion=$(VERSION) -X $(PACKAGE)/core.BuildDate=$(DATE)"

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")


.PHONY: all build clean install uninstall fmt simplify check run release

all: $(TARGET)

$(TARGET): $(SRC)
	go build -ldflags $(LDFLAGS) -o $(TARGET) cmd/mapago/mapago.go

XYZs = windows:amd64 linux:arm64

release:

	@rm -rf release/
	@echo "build release for common architecturs"
	@echo "release tag: $(VERSION)"
	@sleep 3s

	# @$(foreach XYZ,$(XYZs), \
	# 	$(eval OS = $(word 1,$(subst :, ,$(XYZ)))) \
	# 	$(eval PLATFORM = $(word 2,$(subst :, ,$(XYZ)))) \
	# 	$(eval  ARCH=release/$(OS)-$(PLATFORM)) \
	# 	- @echo "build $(OS) Executable for $(PLATFORM)\n" \
	# 	@rm -rf $(ARCH) \
	# 	${MKDIR_P} $(ARCH) \
	# 	GOOS=$(OS) GOARCH=$(PLATFORM) go build -ldflags $(LDFLAGS) -o $(ARCH)/$(TARGET) cmd/mapago/mapago.go \
	# )

	$(eval OS := windows)
	$(eval PLATFORM := amd64)
	$(eval  ARCH=release/$(OS)-$(PLATFORM))
	@echo "build $(OS) Executable for $(PLATFORM)"
	@rm -rf $(ARCH)
	${MKDIR_P} $(ARCH)
	GOOS=$(OS) GOARCH=$(PLATFORM) go build -ldflags $(LDFLAGS) -o $(ARCH)/$(TARGET).exe cmd/mapago/mapago.go

	$(eval OS := linux)
	$(eval PLATFORM := arm64)
	$(eval  ARCH=release/$(OS)-$(PLATFORM))
	@echo "build $(OS) Executable for $(PLATFORM)"
	@rm -rf $(ARCH)
	${MKDIR_P} $(ARCH)
	GOOS=$(OS) GOARCH=$(PLATFORM) go build -ldflags $(LDFLAGS) -o $(ARCH)/$(TARGET) cmd/mapago/mapago.go

	$(eval OS := linux)
	$(eval PLATFORM := amd64)
	$(eval  ARCH=release/$(OS)-$(PLATFORM))
	@echo "build $(OS) Executable for $(PLATFORM)"
	@rm -rf $(ARCH)
	${MKDIR_P} $(ARCH)
	GOOS=$(OS) GOARCH=$(PLATFORM) go build -ldflags $(LDFLAGS) -o $(ARCH)/$(TARGET) cmd/mapago/mapago.go

	$(eval OS := darwin)
	$(eval PLATFORM := amd64)
	$(eval  ARCH=release/$(OS)-$(PLATFORM))
	@echo "build $(OS) Executable for $(PLATFORM)"
	@rm -rf $(ARCH)
	${MKDIR_P} $(ARCH)
	GOOS=$(OS) GOARCH=$(PLATFORM) go build -ldflags $(LDFLAGS) -o $(ARCH)/$(TARGET) cmd/mapago/mapago.go


install:
	go install -ldflags $(LDFLAGS) management-plane/mapago-client.go
	go install -ldflags $(LDFLAGS) management-plane/mapago-server.go
	go install -ldflags $(LDFLAGS) management-plane/spawnTopology.go

build: $(TARGET)
	@true

clean:
	@rm -f $(TARGET)

fmt:
	gofmt -l -w $(SRC)

simplify:
	gofmt -s -l -w $(SRC)

check:
	@test -z $(shell gofmt -l main.go | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
	@go tool vet ${SRC}
