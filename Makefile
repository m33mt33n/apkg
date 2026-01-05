
CC = gcc

all: aay-bin-dynamic-install

aay-bin-dynamic-install: cmd/aay
	go install -ldflags="-s -w" -tags "libsqlite3 linux" -v github.com/m33mt33n/apkg/cmd/aay

aay-bin-dynamic-build: cmd/aay
	go build -ldflags="-s -w" -tags "libsqlite3 linux" -o aay -v ./cmd/aay

aay-bin-dynamic-build-temp: cmd/aay
	go build -ldflags="-s -w" -tags "libsqlite3 linux" -o /dev/shm/aay -v ./cmd/aay

# m3: static (glibc) build showing below warning, as system's (trixie) libsqlite3 is built with support for sqlite extensions,
# binary worked fine in android environment, error could be suppressed by building from source as instructed in mattn/go-sqlite3
# readme.
# /usr/bin/ld: /usr/lib/gcc/arm-linux-gnueabihf/14/../../../arm-linux-gnueabihf/libsqlite3.a(os_unix.o): in function `unixDlOpen':
# (.text+0xaa2): warning: Using 'dlopen' in statically linked applications requires at runtime the shared libraries from the
# glibcversion used for linking
aay-bin-static-glibc-build: cmd/aay
	go build -ldflags="-linkmode external -extldflags '-static -lm'" -tags "libsqlite3 linux" -o aay -v ./cmd/aay

aay-bin-static-glibc-build-temp: cmd/aay
	go build -ldflags="-linkmode external -extldflags '-static -lm'" -tags "libsqlite3 linux" -o /dev/shm/aay -v ./cmd/aay

# m3: below method as suggested by mattn/go-sqlite3 readme, failed with so many errors about undefined references mostly related to sqlite3.
# reason/solution to failure is yet to be found.
#aay-bin-static-musl-build-temp-failed: cmd/aay
	#CC=musl-gcc go build --ldflags '-linkmode external -extldflags=-static -s -w' -tags sqlite_omit_load_extension -o /dev/shm/aay -v ./cmd/aay

# m3: as workaround: below musl tasks need a libsqlite3.a built with musl-gcc.
aay-bin-static-musl-build: cmd/aay
	CC=musl-gcc go build --ldflags '-linkmode external -extldflags=-static -s -w' -tags "libsqlite3 linux sqlite_omit_load_extension" -o aay -v ./cmd/aay

aay-bin-static-musl-build-temp: cmd/aay
	CC=musl-gcc go build --ldflags '-linkmode external -extldflags=-static -s -w' -tags "libsqlite3 linux sqlite_omit_load_extension" -o /dev/shm/aay -v ./cmd/aay

aay-bin-static-musl-run: cmd/aay
	CC=musl-gcc go run --ldflags '-linkmode external -extldflags=-static -s -w' -tags "libsqlite3 linux sqlite_omit_load_extension" -v ./cmd/aay $(AAY_ARGS)

aay-bin-static-musl-build-embed-aapt-temp: cmd/aay
	CC=musl-gcc go build --ldflags '-linkmode external -extldflags=-static -s -w' -tags "libsqlite3 linux sqlite_omit_load_extension embed_aapt" -o /dev/shm/aay -v ./cmd/aay

clean:
	rm -f aay
