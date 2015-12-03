all: compile

compile:
	go build
	go build ./base
	go build ./gpmd

test:
	gocov test ./ ./base ./gpmd | gocov report
