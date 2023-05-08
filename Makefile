# This makefile expects a following folders structure:
# .
# ├─inc #include files, you can override this folder by setting LIBINSANE_GO_INCDIR
# │  └─libinsane 
# │     ├─capi.h
# │     └─...
# └─lib #lib files, you can override this folder by setting LIBINSANE_GO_LIBDIR
#    ├─libinsane32.a #for x86 builds
#    └─libinsane64.a #for x86_64 builds
#   
cur_dir := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
bin_dir = bin
ifeq ($(OS),Windows_NT)
	detected_OS := Windows
	#under mingw cur_dir is in the form /c/path/to/folder that is inacceptable for CGO_CFLAGS -I
	#convert it to c:\path\to\folder format with shell ( echo "/c/path/to/folder/" | sed -r "s#^/(.{1})/#\1:\\\#" ) 
	cur_dir != echo "$(cur_dir)" | sed -r "s=^/(.{1})/=\1:\\\="
	cur_dir := $(subst /,\,$(cur_dir))
	LIBINSANE_GO_INCDIR ?= $(cur_dir)\inc
	LIBINSANE_GO_LIBDIR ?= $(cur_dir)\lib
	os1 = windows
	ifeq ($(MSYSTEM), MINGW32)
		arch = 32
	else
		arch = 64
	endif
else
	detected_OS := $(shell uname)  # same as "uname -s"
	ifeq ($(strip $(detected_OS)),Linux)
		os1 = linux
	else
		ifeq ($(strip $(detected_OS)),Darwin)
			os1 = macos
		else
$(error Unknown OS: $(detected_OS))
		endif
	endif
$(error NOT IMPLEMENTED YET!!!) #not tested on Mac and Linux yet
endif


.PHONY: all build

all: build

build: $(bin_dir)
	@export CGO_CFLAGS="-I $(LIBINSANE_GO_INCDIR)"; \
		export CGO_LDFLAGS="--static $(LIBINSANE_GO_LIBDIR)\libinsane$(arch).a -lregex -lole32  -loleaut32 -luuid -lsystre -ltre  -lpthread -lintl -liconv"; \
		go build -o $(bin_dir)/lisgo$(arch).exe cmd/lisgo/*.go

$(bin_dir):
	@mkdir $(bin_dir)


