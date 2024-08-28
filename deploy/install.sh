#! /bin/bash

# 检测是否安装了wget
if ! command -v wget >/dev/null 2>&1; then
  echo "wget is not installed, please install it first."
  exit 1
fi

# 检测系统架构和操作系统
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# 根据架构和操作系统确定下载链接
if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
  echo "Unsupported OS: $OS"
  exit 1
fi

#创建目录
mkdir -p /root/data/doh-client

#获取最新版本号
VERSION=$(curl -s https://raw.githubusercontent.com/WJQSERVER/doh-client/main/VERSION)
if [ -z "$VERSION" ]; then
  echo "获取最新版本号失败，请检查网络连接或稍后再试"
  exit 1
fi
echo "最新版本：$VERSION"

#下载最新版本的doh-client
wget -O /root/data/doh-client.tar.gz https://github.com/WJQSERVER/doh-client/releases/download/$VERSION/doh-client-$OS-$ARCH.tar.gz

#解压并移动到指定位置
tar -zxvf /root/data/doh-client.tar.gz -C /root/data/doh-client
mv /root/data/doh-client/doh-client-$OS-$ARCH /root/data/doh-client/doh-client
rm -rf /root/data/doh-client.tar.gz

#设置权限
chmod +x /root/data/doh-client/doh-client
chown root:root /root/data/doh-client/doh-client

#拉取配置文件
mkdir -p /root/data/doh-client/config
mkdir -p /root/data/doh-client/log
#检测配置是否已存在
if [ ! -f /root/data/doh-client/config/config.yaml ]; then
  wget -O /root/data/doh-client/config/config.yaml https://raw.githubusercontent.com/WJQSERVER/doh-client/main/config/config.yaml
fi

#拉取systemd service文件
wget -O /etc/systemd/system/doh-client.service https://raw.githubusercontent.com/WJQSERVER/doh-client/main/deploy/doh-client.service

#启动服务
systemctl enable doh-client
systemctl start doh-client
