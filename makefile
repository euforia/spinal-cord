
NURV=nurv
SPINAL_CORD=spinal-cord

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

BUILD_DIR="build"
INSTALL_DIR="/usr/local/bin/"

all: preflight nurv spinal-cord

preflight:
	[ -e ${BUILD_DIR} ] || mkdir ${BUILD_DIR}

nurv: preflight
	cd ${BUILD_DIR} && go build -o ${NURV}.${GOOS}_${GOARCH} ../${NURV}/main.go

spinal-cord: preflight
	cd ${BUILD_DIR} && go build -o ${SPINAL_CORD}.${GOOS}_${GOARCH} ../${SPINAL_CORD}/main.go

clean:
	rm -rvf ${BUILD_DIR}

install: all
	cp ${BUILD_DIR}/nurv ${INSTALL_DIR}
	cp ${BUILD_DIR}/spinal-cord ${INSTALL_DIR}