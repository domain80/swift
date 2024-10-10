WINDOWS_DIR=./dist/windows/swift_x64
WINDOWS86_DIR=./dist/windows/swift_x86

INTEL_MAC_DIR=./dist/macOS_intel/swift_x64
APPLE_SILICON=./dist/macOS_appleSilicon/swift_x64

LINUX_DIR=./dist/linux/swift_x64
LINUX86_DIR=./dist/linux/swift_x86

# windows
# 86-bit
GOOS=windows GOARCH=386 go  build  -ldflags="-w -s" -gcflags=all=-l -o $WINDOWS86_DIR main.go
7z a -tzip $WINDOWS_DIR.zip $WINDOWS_DIR.exe 

# 64-bit
GOOS=windows GOARCH=amd64 go  build  -ldflags="-w -s" -gcflags=all=-l -o $WINDOWS_DIR main.go
7z a -tzip $WINDOWS_DIR.zip $WINDOWS_DIR.exe 

# macos
# 64-bit
GOOS=darwin GOARCH=amd64 go  build  -ldflags="-w -s" -gcflags=all=-l -o $INTEL_MAC_DIR main.go
7z a -tzip $INTEL_MAC_DIR.zip $INTEL_MAC_DIR

GOOS=darwin GOARCH=arm64 go  build  -ldflags="-w -s" -gcflags=all=-l -o $APPLE_SILICON main.go
7z a -tzip $APPLE_SILICON.zip $APPLE_SILICON 

# # linux
# # 64-bit
GOOS=linux GOARCH=amd64 go  build  -ldflags="-w -s" -gcflags=all=-l -o $LINUX_DIR main.go
7z a -tzip $LINUX_DIR.zip $LINUX_DIR 

# # 86-bit
GOOS=linux GOARCH=386 go  build -ldflags="-w -s" -gcflags=all=-l  -o $LINUX86_DIR main.go
7z a -tzip $LINUX86_DIR.zip $LINUX86_DIR 

# rm ./dist/linux/swift-386-linux
# rm ./dist/linux/swift-amd64-linux
# rm ./dist/macos/swift-amd64-darwin
# rm ./dist/windows/swift-386.exe
# rm ./dist/windows/swift-amd64.exe

rm $WINDOWS_DIR.exe
rm $WINDOWS86_DIR.exe
rm $INTEL_MAC_DIR
rm $APPLE_SILICON
rm $LINUX_DIR
rm $LINUX86_DIR

