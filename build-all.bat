@echo off
echo Building for all platforms...

set GOOS=windows& set GOARCH=amd64& go build -o server-eye-windows-amd64.exe
echo ✅ Windows 64-bit done

set GOOS=windows& set GOARCH=386& go build -o server-eye-windows-386.exe
echo ✅ Windows 32-bit done

set GOOS=linux& set GOARCH=amd64& go build -o server-eye-linux-amd64
echo ✅ Linux 64-bit done

set GOOS=linux& set GOARCH=arm64& go build -o server-eye-linux-arm64
echo ✅ Linux ARM64 (Raspberry Pi) done

set GOOS=darwin& set GOARCH=amd64& go build -o server-eye-macos-intel
echo ✅ macOS Intel done

set GOOS=darwin& set GOARCH=arm64& go build -o server-eye-macos-apple-silicon
echo ✅ macOS Apple Silicon done

echo 🎉 All builds complete!
dir server-eye-*