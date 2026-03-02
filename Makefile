.PHONY: dev build

# On Linux, fall back to webkit2gtk-4.1 if 4.0 is not available (e.g. Ubuntu 24.04+)
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
  ifeq ($(shell pkg-config --exists webkit2gtk-4.0 && echo yes || echo no),no)
    TAGS := -tags webkit2_41
  endif
endif

dev:
	wails dev $(TAGS)

build:
	wails build $(TAGS)
