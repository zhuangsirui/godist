LIBS = ./ ./base ./gpmd

all: compile

compile:
	go build ${LIBS}
	go install ./gpmd_server

test:
	gocov test ${LIBS} | gocov report

deps:
	go get -u github.com/smartystreets/goconvey/convey
