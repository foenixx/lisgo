# lisgo
Go bindings for [*libinsane*](https://gitlab.gnome.org/World/OpenPaperwork/libinsane) scanning library (OpenPaperwork project).

**lisgo** library is potentially cross-platform, but tested only on Windows yet. 

## How to compile on Windows

Instructions for compiling 32-bit version.

1. Install MSYS2 and everything for compiling 32-bit programs. Run it in MINGW32 shell. This is an excerpt from [libinsane documentation](https://doc.openpaper.work/libinsane/latest/libinsane/install.html). 
    ```
    pacman -Syuu
    pacman -S \
        make \
        mingw-w64-i686-cunit \
        mingw-w64-i686-doxygen \
        mingw-w64-i686-gcc \
        mingw-w64-i686-gobject-introspection \
        mingw-w64-i686-meson \
        mingw-w64-i686-python3-gobject \
        mingw-w64-i686-vala
    ```
    If you plan on compiling your program as a 64bits executable, you just have to replace all the `i686` by `x86_64` and run it in MINGW64 shell.

2. Download libinsane sources
    ```
    wget -c https://gitlab.gnome.org/World/OpenPaperwork/libinsane/-/archive/1.0.10/libinsane-1.0.10.tar.gz -O - | tar -xz
    mv libinsane-* libinsane
    ```
3. To compile libinsane as a static library, change line #88 in `subprojects\libinsane\src\meson.build` from `libinsane = library` to `libinsane = static_library`
    ```
    (cd libinsane && sed -i 's/libinsane = library/libinsane = static_library/g' subprojects/libinsane/src/meson.build)
    ```
4. Build libinsane
    ```
    (cd libinsane && make PREFIX=/mingw64)
    ```
5. Set CGO options
    ```
    export LIBINSANE_DIR=`cygpath -aw libinsane`
    export CGO_CFLAGS="-I${LIBINSANE_DIR}/subprojects/libinsane/include"
    export CGO_LDFLAGS="-L${LIBINSANE_DIR}/build/subprojects/libinsane/src/ -static -linsane -lpthread -lsystre -lintl -ltre -liconv -lregex -lole32 -loleaut32 -luuid"
6. Build
    ```
    go get .
    go build -o bin/lisgo.exe cmd/lisgo/main.go
    ```

## How to compile on Ubuntu

1. Install dependencies
    ```
    sudo apt get update
    sudo apt install build-essential libinsane-dev libsane-dev
    ```
2. Set CGO options
    ```
    export CGO_LDFLAGS="-linsane"
    ```
3. Build
    ```
    go get .
    go build -o bin/lisgo cmd/lisgo/main.go
    ```

## lisgo.exe command-line utility

This project includes `lisgo.exe` command-line utility. It illustrates using of the library. Please refer to `cmd\lisgo\lisgo.go` for examples.
`lisgo.exe` could be compiled as 32-bit or 64-bit program. Usually 32-bit is preferred way of using it, because most of Twain drivers are 32-bit only.

Currently lisgo.exe can:

* List available scanners.
```
lisgo32.exe print-scanners

Qt: Untested Windows version 10.0 detected!
---------------------------------------------------------------------
Device Id: twain:TWAIN Working Group:TWAIN2 FreeImage Software Scanner

Vendor: TWAIN Working Group
Model: TWAIN2 FreeImage Software Scanner
Qt: Untested Windows version 10.0 detected!
Qt: Untested Windows version 10.0 detected!
Paper source: flatbed
Paper source: feeder
---------------------------------------------------------------------
Device Id: twain:Brother Industries, Ltd.:TW-Brother MFC-L3770CDW LAN

Vendor: Brother Industries, Ltd.
Model: TW-Brother MFC-L3770CDW LAN
   ⨯ libinsane error           error=../subprojects/libinsane/src/bases/twain/twain.c:L1551(twain_simple_set_value): Brother Industries, Ltd.:TW-Brother MFC-L3770CDW LAN->simple_set_value(supported_sizes): Failed to get value: 0x60000002, LibInsane internal error: Unknown error reported by backend (please report !)
   ⨯ libinsane error           error=../subprojects/libinsane/src/workarounds/cache.c:L229(cache_set_value): supported_sizes->set_value() failed: 0x60000002, LibInsane internal error: Unknown error reported by backend (please report !)
Paper source: feeder
Paper source: flatbed
```
* Print specific scanner and paper source options.
```
lisgo32.exe print-options -d "twain:Brother Industries, Ltd.:TW-Brother MFC-L3770CDW LAN" -s feeder

------- transfer_count ------
transfer_count (transfer_count;transfer_count)
Caps: 8 [LisCapSwSelect]
Type: 1 (Integer)
Units: 0 (None)
Constraint: List (-1)
Value: -1

------- compression ------
compression (compression;compression)
Caps: 8 [LisCapSwSelect]
Type: 3 (String)
Units: 0 (None)
Constraint: List (none)
Value: none

------- mode ------
mode (mode;mode)
Caps: 8 [LisCapSwSelect]
Type: 3 (String)
Units: 0 (None)
Constraint: List (LineArt,Gray,Color)
Value: Color

------- units ------
units (units;units)
Caps: 8 [LisCapSwSelect]
Type: 3 (String)
Units: 0 (None)
Constraint: List (inches,centimeters,pixels)
Value: inches
...
```
* Scan to pdf, jpeg and png format. Default format is pdf, resulting file name is result.pdf. 
```
lisgo32.exe scan -d "twain:Brother Industries, Ltd.:TW-Brother MFC-L3770CDW LAN" -s feeder
```
You can use `-o` flag to set scanner options. Flag `-o` can be specified more than once to set several options.
For example, for duplex gray-scale scanning, you can issue this command:
```
lisgo32.exe scan -o mode=Gray -o duplex_enabled=true -d "twain:Brother Industries, Ltd.:TW-Brother MFC-L3770CDW LAN" -s feeder`
```
* Print help page for available commands.
```
lisgo32.exe scan

Scan using specified scanner and paper source. Output file will have name like 'page1.png, page2.jpg or result.pdf' depending on -f option value.

Options:
  -d string
        id of the scanner, mandatory
  -f string
        output file format [jpg|png|pdf] (default "pdf")
  -o value
        try to set specified option before scan.
        Format:
        -o name=value :  set option with [name] to [value]
        -o name= : pass empty string as value of the option
        This flag can appear multiple times: -o name1=value1 -o name2=value2
  -s string
        paper source, mandatory
  -v    show debug messages
```

