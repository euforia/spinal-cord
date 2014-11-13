
NURV=nurv
SPINAL_CORD=spinal-cord

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

BUILD_DIR="build"

all: preflight nurv spinalcord

preflight:
	[ -e ${BUILD_DIR} ] || mkdir ${BUILD_DIR}

nurv: preflight
	cd ${BUILD_DIR} && go build -o ${NURV}.${GOOS}_${GOARCH} ../${NURV}.go

spinal-cord: preflight
	cd ${BUILD_DIR} && go build -o ${SPINAL_CORD}.${GOOS}_${GOARCH} ../${SPINAL_CORD}.go

clean:
	rm -rvf ${BUILD_DIR}
