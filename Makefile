build-windows:
	cd src && \
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="x86_64-w64-mingw32-gcc" go build -ldflags "-w -s  -H=windowsgui" -o ../bin/iserverlookup.exe -mod=readonly && \
	cd ..\bin && \
	osslsigncode sign -pkcs12 ~/Dropbox/swap/golang/convergelookup/fyne-cross/dist/windows-arm64/codesign.pfx -pass "nopasswordforyou" -t http://timestamp.digicert.com -in iserverlookup.exe -out signed-iserverlookup.exe

build-osx:
	cd src && \
	GOFLAGS="-ldflags=-extldflags=-Wl,-ld_classic,-no_warn_duplicate_libraries,-v" fyne package -os darwin 

build-oldosx:
	cd src && \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CGO_CFLAGS="-mmacosx-version-min=10.12" CGO_LDFLAGS="-mmacosx-version-min=10.12" go build -mod=readonly -o ../bin/iserverlookupold

build-newosx:
	cd src && \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 CGO_CFLAGS="-mmacosx-version-min=10.14" CGO_LDFLAGS="-mmacosx-version-min=10.14" go build -mod=readonly -o ../bin/iserverlookupnew

build-osxu: build-oldosx build-newosx build-osx
	cd bin && \
	lipo -create -output iserverlookup iserverlookupold iserverlookupnew && \
	cp iserverlookup ../src/iServer\ Lookup.app/Contents/MacOS && \
	codesign -f -s - ../src/iServer\ Lookup.app