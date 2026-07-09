#!/bin/bash

echo "Assert that script is not run by root user"
if [ "$(id -u)" = "0" ]; then
    echo "This script should not be run as root" 1>&2
    exit 1
fi

echo "Installing required debian packages"
sudo apt-get update
sudo apt-get install -y wget git curl

echo "Installing docker"
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER

echo "Installing ARM64 Docker build emulation"
# Needed by maintainers who build arm64 Docker images on amd64 hosts.
sudo docker run --privileged --rm tonistiigi/binfmt --install arm64

echo "Installing go"
goVersion="1.26.0"
goArchive="go${goVersion}.linux-amd64.tar.gz"
goUrl="https://go.dev/dl/${goArchive}"
wget -O "$goArchive" "$goUrl"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "$goArchive"
rm "$goArchive"
bashrcPath="$HOME/.bashrc"
pathLine='export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"'
grep -qxF "$pathLine" "$bashrcPath" || echo "$pathLine" >> "$bashrcPath"

echo "Preparing .bashrc"
source $HOME/.bashrc

echo "Configuring Go toolchain auto-upgrade"
go env -w GOTOOLCHAIN=go${goVersion}+auto

echo "Configuring Linux user namespace settings for sandboxed browser acceptance tests"
# Rod acceptance tests launch Chromium with sandbox enabled. On some Ubuntu/AppArmor setups,
# unprivileged user namespaces are blocked by default, which causes Chromium startup to fail with
# "No usable sandbox". This configuration is required for local developer acceptance tests.
# Review this against your organization security policy before using in hardened environments.
sudo tee /etc/sysctl.d/99-quollix-browser-sandbox.conf >/dev/null <<'EOF'
kernel.unprivileged_userns_clone=1
user.max_user_namespaces=28633
kernel.apparmor_restrict_unprivileged_userns=0
EOF
sudo sysctl --system

echo "Please reboot the system to complete the installation."
