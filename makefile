LIBS = godist godist/gpmd

all: compile

compile:
	go build ${LIBS}

test:
	go get github.com/axw/gocov/gocov
	gocov test ${LIBS} | gocov report
