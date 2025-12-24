# Hướng dẫn cài đặt Gitea (Git Server) trên Ubuntu ARM64 (Parallels)

Tài liệu này ghi lại các bước đã thực hiện để triển khai Git Server nội bộ cho team.

## 1. Thông tin hệ thống
- **OS:** Ubuntu 24.04 LTS (ARM64) chạy trên Parallels Desktop (macOS).
- **Phần mềm:** Gitea (viết bằng Go, nhẹ và giao diện giống GitHub).
- **Port truy cập:** 999 (đã cấu hình đặc biệt).

## 2. Các bước cài đặt đã thực hiện

### Bước 1: Chuẩn bị User và Thư mục
```bash
# Tạo user hệ thống cho Git
sudo adduser --system --shell /bin/bash --gecos 'Git Version Control' --group --disabled-password --home /home/git git

# Tạo cấu trúc thư mục dữ liệu
sudo mkdir -p /var/lib/gitea/{custom,data,log}
sudo chown -R git:git /var/lib/gitea/
sudo chmod -R 750 /var/lib/gitea/

# Tạo thư mục cấu hình
sudo mkdir -p /etc/gitea
sudo chown root:git /etc/gitea
sudo chmod 770 /etc/gitea
```

### Bước 2: Cài đặt Binary Gitea
```bash
wget -O gitea https://dl.gitea.com/gitea/1.21.1/gitea-1.21.1-linux-arm64
sudo chmod +x gitea
sudo mv gitea /usr/local/bin/gitea

# Cấp quyền cho phép Gitea chạy trên port thấp (<1024)
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/gitea
```

### Bước 3: Cấu hình Systemd Service
File: `/etc/systemd/system/gitea.service`
```ini
[Unit]
Description=Gitea (Git with a cup of tea)
After=network.target

[Service]
RestartSec=2s
Type=simple
User=git
Group=git
WorkingDirectory=/var/lib/gitea/
ExecStart=/usr/local/bin/gitea web --config /etc/gitea/app.ini
Restart=always
Environment=USER=git HOME=/home/git GITEA_WORK_DIR=/var/lib/gitea

[Install]
WantedBy=multi-user.target
```

### Bước 4: Cấu hình Port 999 và Firewall
```bash
# Mở port trên firewall
sudo ufw allow 999/tcp

# Kích hoạt dịch vụ
sudo systemctl enable --now gitea
```

## 3. Cấu hình Web ban đầu
Truy cập: `http://<IP_MAY_AO>:999`

**Thông số quan trọng:**
- **Database Type:** SQLite3
- **HTTP Port:** 999
- **Base URL:** `http://<IP_MAY_AO>:999/`
- **Admin Account:** Cần tạo trong mục *Optional Settings* khi cài đặt lần đầu.

## 4. Lệnh quản lý nhanh
- Kiểm tra trạng thái: `sudo systemctl status gitea`
- Khởi động lại: `sudo systemctl restart gitea`
- Xem log trực tiếp: `sudo journalctl -u gitea -f`
- Sửa cấu hình trực tiếp: `sudo nano /etc/gitea/app.ini`
