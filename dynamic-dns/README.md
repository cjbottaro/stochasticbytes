## Dynamic DNS

This periodically queries our public IP address and updates the
DNS record in dnsimple.

I wrote this in Rust for fun. There is a Makefile to build and push it to Docker Hub.
```sh
make
```

Linking to openssl in Alpine causes segfaults unless you use this flag:
```
RUSTFLAGS=-C target-feature=-crt-static
```
https://users.rust-lang.org/t/sigsegv-with-program-linked-against-openssl-in-an-alpine-container/52172

I tried to be a statically linked binary with this flag:
```
RUSTFLAGS=-C link-arg=-s
```
https://stackoverflow.com/questions/59766239/how-to-build-a-rust-app-free-of-shared-libraries

But it doesn't work. I still have to install `libgcc` on the final image.