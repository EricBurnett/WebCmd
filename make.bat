mkdir ..\..\..\..\bin
call go build -o WebCmd.exe -ldflags="-Hwindowsgui" main.go server.go modulelist.go
copy WebCmd.exe ..\..\..\..\bin\WebCmd.exe