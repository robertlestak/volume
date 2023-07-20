bin: bin/webfs_darwin_amd64 bin/webfs_darwin_arm64 bin/webfs_linux_amd64 
bin: bin/webfs_linux_arm bin/webfs_linux_arm64

bin/webfs_darwin_amd64:
	GOOS=darwin GOARCH=amd64 go build -o bin/webfs_darwin_amd64 cmd/webfs/*.go

bin/webfs_darwin_arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/webfs_darwin_arm64 cmd/webfs/*.go

bin/webfs_linux_amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/webfs_linux_amd64 cmd/webfs/*.go

bin/webfs_linux_arm:
	GOOS=linux GOARCH=arm go build -o bin/webfs_linux_arm cmd/webfs/*.go

bin/webfs_linux_arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/webfs_linux_arm64 cmd/webfs/*.go

.PHONY: bin