#/bin/sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o mybinary_windows.exe -v ./src/main.go;
echo "windows可执行文件生成成功";
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o mybinary_mac -v ./src/main.go;
echo "mac可执行文件生成成功"

echo "success！"

