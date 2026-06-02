```powershell
# PowerShell — ARCH = amd64 or arm64
$Arch = "amd64"
Invoke-WebRequest `
  -Uri "https://github.com/garenal1v3/tgs-cli/releases/latest/download/tgs-cli_windows_$Arch.zip" `
  -OutFile tgs.zip
Expand-Archive tgs.zip -DestinationPath .
# Move tgs.exe into a directory on your PATH, then:
.\tgs.exe version
```
