LIBS = ./ ./base ./gpmd

all: compile

compile:
	go build ${LIBS}

test:
	gocov test ${LIBS} | gocov report
