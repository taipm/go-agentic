# Hướng dẫn cài đặt Gitea (Git Server) trên Ubuntu ARM64 (Parallels)

Tài liệu này ghi lại các bước đã thực hiện để triển khai Git Server nội bộ cho team.

## 1. Thông tin hệ thống
- **OS:** Ubuntu 24.04 LTS (ARM64) chạy trên Parallels Desktop (macOS).
- **Phần mềm:** Gitea (viết bằng Go, nhẹ và giao diện giống GitHub).
- **Port truy cập:** 3999 (đã chuyển từ 999).
- **Domain:** https://git.microai.club (qua Cloudflare Tunnel).

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

# Cấp quyền cho phép Gitea chạy trên port thấp (<1024) nếu cần
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

### Bước 4: Cấu hình Port 3999 và Firewall
```bash
# Mở port trên firewall
sudo ufw allow 3999/tcp

# Kích hoạt dịch vụ
sudo systemctl enable --now gitea
```

### Bước 5: Cài đặt Cloudflare Tunnel (cloudflared)
```bash
# Cài đặt curl nếu chưa có
sudo apt update && sudo apt install curl -y

# Thêm GPG key và Repo
curl -fsSL https://pkg.cloudflare.com/cloudflare-public-v2.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloudflare-main.gpg
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared any main' | sudo tee /etc/apt/sources.list.d/cloudflared.list

# Cài đặt cloudflared
sudo apt update && sudo apt install cloudflared -y

# Cài đặt service với Token từ Dashboard
sudo cloudflared service install <YOUR_TOKEN>
```

## 3. Cấu hình Web và Cloudflare
Truy cập: `http://localhost:3999` (nội bộ) hoặc `https://git.microai.club` (ngoài).

**Thông số quan trọng trong `/etc/gitea/app.ini`:**
- **HTTP_ADDR:** 0.0.0.0
- **HTTP_PORT:** 3999
- **DOMAIN:** git.microai.club
- **ROOT_URL:** https://git.microai.club/
- **SSH_DOMAIN:** git.microai.club

**Cấu hình trên Cloudflare Zero Trust Dashboard:**
- **Public Hostname:** git.microai.club
- **Service:** HTTP://localhost:3999

## 4. Lệnh quản lý nhanh
- Kiểm tra Gitea: `sudo systemctl status gitea`
- Kiểm tra Tunnel: `sudo systemctl status cloudflared`
- Xem log Tunnel: `sudo journalctl -u cloudflared -f`
- Sửa cấu hình Gitea: `sudo nano /etc/gitea/app.ini`

## 5. Lưu ý quan trọng (Troubleshooting)
- **Lỗi 502:** Thường do Cloudflare Tunnel không kết nối được tới port của Gitea. Kiểm tra lại port trong Dashboard Cloudflare và `app.ini`.
- **Lỗi chính tả:** Kiểm tra kỹ `ROOT_URL` trong `app.ini`. Một lỗi nhỏ như `lcub` thay vì `club` sẽ khiến Gitea không thể tải các tài nguyên (CSS/JS) hoặc redirect sai.
- **Restart:** Luôn chạy `sudo systemctl restart gitea` sau khi sửa file `app.ini`.
- Xem log Tunnel: `sudo journalctl -u cloudflared -f`
- Sửa cấu hình Gitea: `sudo nano /etc/gitea/app.ini`
