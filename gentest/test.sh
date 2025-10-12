go run ../cmd/vera.go -f config-test.dbc .
cd build
cmake ..
make
./test
