.PHONY: clean test

test: build
	./test

clean:
	rm -rf vera* *.o test

build: pre-build test.o vera.o unity.o
	cc -o test test.o vera.o unity.o
	
test.o: test.c
	cc -c test.c

vera.o: vera.c vera.h
	cc -c vera.c

unity.o: unity/unity.c unity/unity.h unity/unity_internals.h
	cc -c unity/unity.c

pre-build: config-test.dbc
	go run ../cmd/vera/main.go -f config-test.dbc .
