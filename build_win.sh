cat > PrettyMath.rc << EOL
id ICON "./Gallery/Icon.ico"
GLFW_ICON ICON "./Gallery/Icon.ico"
EOL

x86_64-w64-mingw32-windres PrettyMath.rc -O coff -o PrettyMath.syso

GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ HOST=x86_64-w64-mingw32 go build -ldflags "-s -w -H=windowsgui -extldflags=-static" -p 4 -v -o PrettyMath.exe

rm PrettyMath.syso
rm PrettyMath.rc
