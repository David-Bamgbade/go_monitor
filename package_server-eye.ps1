# package.ps1 - Simple packaging script for Server Eye

Write-Host "Packaging Server Eye..." -ForegroundColor Green

# Create folders
New-Item -ItemType Directory -Force -Path "server-eye-dist" | Out-Null
New-Item -ItemType Directory -Force -Path "server-eye-dist\windows" | Out-Null
New-Item -ItemType Directory -Force -Path "server-eye-dist\linux" | Out-Null
New-Item -ItemType Directory -Force -Path "server-eye-dist\macos" | Out-Null
New-Item -ItemType Directory -Force -Path "server-eye-dist\arm" | Out-Null

Write-Host "Folders created" -ForegroundColor Green

# Copy files
Copy-Item "server-eye-windows-amd64.exe" -Destination "server-eye-dist\windows\server-eye-64bit.exe"
Copy-Item "server-eye-windows-386.exe" -Destination "server-eye-dist\windows\server-eye-32bit.exe"
Copy-Item "server-eye-linux-amd64" -Destination "server-eye-dist\linux\server-eye"
Copy-Item "server-eye-linux-arm64" -Destination "server-eye-dist\arm\server-eye"
Copy-Item "server-eye-macos-intel" -Destination "server-eye-dist\macos\server-eye-intel"
Copy-Item "server-eye-macos-apple-silicon" -Destination "server-eye-dist\macos\server-eye-m1"

Write-Host "Files copied" -ForegroundColor Green

# Create README files
@"
# SERVER EYE

CPU Spike Detective

Download: https://your-site.com

Contact: davidbamgbade082@gmail.com
"@ | Out-File -FilePath "server-eye-dist\README.md" -Encoding UTF8

@"
SERVER EYE - Windows
Run: server-eye-64bit.exe
Contact: davidbamgbade082@gmail.com
"@ | Out-File -FilePath "server-eye-dist\windows\README.txt" -Encoding UTF8

@"
SERVER EYE - Linux
Run: chmod +x server-eye && ./server-eye
Contact: davidbamgbade082@gmail.com
"@ | Out-File -FilePath "server-eye-dist\linux\README.txt" -Encoding UTF8

@"
SERVER EYE - macOS
Run: chmod +x server-eye-intel && ./server-eye-intel
Contact: davidbamgbade082@gmail.com
"@ | Out-File -FilePath "server-eye-dist\macos\README.txt" -Encoding UTF8

@"
SERVER EYE - ARM/RPi
Run: chmod +x server-eye && ./server-eye
Contact: davidbamgbade082@gmail.com
"@ | Out-File -FilePath "server-eye-dist\arm\README.txt" -Encoding UTF8

Write-Host "README files created" -ForegroundColor Green

# Create ZIP files
Compress-Archive -Path "server-eye-dist\*" -DestinationPath "server-eye-complete.zip" -Force

Write-Host ""
Write-Host "DONE! File created: server-eye-complete.zip" -ForegroundColor Green
Write-Host "Contact: davidbamgbade082@gmail.com" -ForegroundColor Cyan