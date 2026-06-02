```bash
# Choose your platform: OS = darwin (macOS) or linux; ARCH = arm64 or amd64
OS=darwin
ARCH=arm64

curl -L -o tgs.tar.gz \
  "https://github.com/garenal1v3/tgs-cli/releases/latest/download/tgs-cli_${OS}_${ARCH}.tar.gz"
tar -xzf tgs.tar.gz tgs
sudo mv tgs /usr/local/bin/
tgs version
```
