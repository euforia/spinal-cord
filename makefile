NURV=nurv
SPINAL_CORD=spinal-cord
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

all:
	cd bin && go build -o ${NURV}.${GOOS}_${GOARCH} ../${NURV}.go
	cd bin && go build -o ${SPINAL_CORD}.${GOOS}_${GOARCH} ../${SPINAL_CORD}.go
