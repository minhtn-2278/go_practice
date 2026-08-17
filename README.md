# Employee Management System

Go thuần với kiến trúc `Handler → Service → Repository`, sử dụng PostgreSQL làm storage.

## Chạy PostgreSQL

```bash
cp .env.example .env
docker compose up -d postgres
```

Migration trong `migrations/` được PostgreSQL image chạy tự động khi khởi tạo volume lần đầu.

## Chạy application

```bash
go mod tidy
go run ./cmd/app
```

Mặc định application chạy tại `http://localhost:8080`.

```bash
curl http://localhost:8080/healthz
```

## Kiểm tra

```bash
gofmt -w .
go test ./...
go test -race ./...
```
