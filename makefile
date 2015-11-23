LIBS = godist

all: compile

compile:
	go get ${LIBS}
	go build ${LIBS}

test:
	go get ${LIBS}
	go test ${LIBS}
