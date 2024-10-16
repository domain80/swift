WINDOWS_DIR=./dist/windows/swift_x64
WINDOWS86_DIR=./dist/windows/swift_x86

INTEL_MAC_DIR=./dist/macOS_intel/swift_x64
APPLE_SILICON=./dist/macOS_appleSilicon/swift_x64

LINUX_DIR=./dist/linux/swift_x64
LINUX86_DIR=./dist/linux/swift_x86

### remove all past files
rm -r ./dist

# windows
# 86-bit
GOOS=windows GOARCH=386 go  build  -ldflags="-w -s" -gcflags=all=-l -o $WINDOWS86_DIR.exe main.go

# 64-bit
GOOS=windows GOARCH=amd64 go  build  -ldflags="-w -s" -gcflags=all=-l -o $WINDOWS_DIR.exe main.go

# macos
# 64-bit
GOOS=darwin GOARCH=amd64 go  build  -ldflags="-w -s" -gcflags=all=-l -o $INTEL_MAC_DIR main.go

GOOS=darwin GOARCH=arm64 go  build  -ldflags="-w -s" -gcflags=all=-l -o $APPLE_SILICON main.go

# # linux
# # 64-bit
GOOS=linux GOARCH=amd64 go  build  -ldflags="-w -s" -gcflags=all=-l -o $LINUX_DIR main.go

# # 86-bit
GOOS=linux GOARCH=386 go  build -ldflags="-w -s" -gcflags=all=-l  -o $LINUX86_DIR main.go

##### zip up files
# 7z a -tzip $WINDOWS_DIR.zip $WINDOWS_DIR.exe 
# 7z a -tzip $WINDOWS_DIR.zip $WINDOWS_DIR.exe 
# 7z a -tzip $INTEL_MAC_DIR.zip $INTEL_MAC_DIR
# 7z a -tzip $APPLE_SILICON.zip $APPLE_SILICON 
# 7z a -tzip $LINUX_DIR.zip $LINUX_DIR 
# 7z a -tzip $LINUX86_DIR.zip $LINUX86_DIR 

##### remove executables and maintain zipped files
# rm $WINDOWS_DIR.exe
# rm $WINDOWS86_DIR.exe
# rm $INTEL_MAC_DIR
# rm $APPLE_SILICON
# rm $LINUX_DIR
# rm $LINUX86_DIR

