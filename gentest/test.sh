go run ../cmd/vera/main.go -f config-test.dbc .
cd build
cmake ..
make
./test
