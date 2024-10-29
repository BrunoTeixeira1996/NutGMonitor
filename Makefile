SHELL := /bin/bash
FILES = nutgmonitor email_template.html
REMOTE_USER = brun0
REMOTE_HOST = pinute
REMOTE_PATH = /home/$(REMOTE_USER)/src/nutgmonitor
BINARY_NAME = nutgmonitor
TARGET_OS = linux
TARGET_ARCH = arm64

compile:
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -o $(BINARY_NAME) ./cmd/nutgmonitor

run-in-ssh:
	$(MAKE) compile
	rsync -avz --update $(FILES) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)
	ssh $(REMOTE_USER)@$(REMOTE_HOST) 'source .bash_profile; cd $(REMOTE_PATH) && ./$(BINARY_NAME)'

gdb:
	GOOS=$(TARGET_OS) go build -gcflags "all=-N -l" -o $(BINARY_NAME) ./cmd/nutgmonitor
	gdb $(BINARY_NAME)

deploy:
	$(MAKE) compile
	rsync -avz --update $(FILES) $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)