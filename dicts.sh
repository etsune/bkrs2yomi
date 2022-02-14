go build -o bin/bkrs2yomi ./cmd/bkrs2yomi/main.go
rm -r dist
mkdir dist
cd dist
../bin/bkrs2yomi -daily
../bin/bkrs2yomi -daily -extended
../bin/bkrs2yomi -daily -type=1
../bin/bkrs2yomi -daily -type=2
