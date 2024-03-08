
build:  build/pam_usher.so

build/pam_usher.so: cmd/session/*.go internal/pam_usher/*
	mkdir -p build
	CGO_CFLAGS="-g -O2" go build --buildmode=c-shared -o build/pam_usher.so cmd/session/session.go cmd/session/pam.go