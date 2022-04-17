ALPINE_VERSION=3.15
RUST_VERSION=1.60.0

default:
	docker buildx build build \
	-t cjbottaro/dynamic-dns:latest \
	--platform linux/amd64 \
	--build-arg ALPINE_VERSION=${ALPINE_VERSION} \
	--build-arg RUST_VERSION=${RUST_VERSION} \
	--push