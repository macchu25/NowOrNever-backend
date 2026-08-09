# Now or Never - Golang Backend Service

Backend API siêu tốc cho ứng dụng **Now or Never** được viết bằng Golang:
- ⚡ **Framework / Runtime**: Go 1.24+ Standard HTTP Engine
- 📡 **REST Endpoints**:
  - `GET /health` - Health check status
  - `GET /api/v1/stats` - Server telemetry (Goroutines, RAM Allocation, Uptime)
  - `GET /api/v1/data` - Retrieve items/activities
  - `POST /api/v1/data` - Create item/activity
- 🐳 **Docker Multi-stage build**: Minimal Alpine container (~15MB image size)

---

## 🚀 Hướng dẫn khởi chạy Local

```bash
# 1. Chạy dịch vụ trực tiếp
go run main.go

# 2. Hoặc biên dịch file thực thi
go build -o server main.go
./server
```

Dịch vụ chạy tại: `http://localhost:8080`
