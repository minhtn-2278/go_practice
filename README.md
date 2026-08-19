# Employee Management System

Ứng dụng quản lý nhân viên viết bằng Go, sử dụng kiến trúc:

```
Handler → Service → Repository → PostgreSQL
```

## Yêu cầu

- Go 1.26+
- Docker và Docker Compose
- PostgreSQL 17 nếu không chạy bằng Docker Compose

## Cấu hình môi trường

Tạo file `.env` từ file mẫu:

```
cp .env.example .env
```

Docker Compose tự đọc `.env`, nhưng `go run` không tự đọc file này. Export các biến trước khi chạy application:

```
set -a
source .env
set +a
```

Kiểm tra connection string:

```
printenv DATABASE_URL
```

Với PostgreSQL local, `DATABASE_URL` cần có `sslmode=disable`.

## Chạy PostgreSQL

```
docker compose up -d go-postgres
```

Các migration trong `migrations/` được PostgreSQL chạy tự động khi khởi tạo volume lần đầu.

Adminer có thể truy cập tại `http://localhost:8081`.

## Chạy application

```
go mod tidy
set -a
source .env
set +a
go run ./cmd/app
```

Application mặc định chạy tại `http://localhost:8080`.

Tất cả endpoint đều yêu cầu Basic Auth. Tài khoản mặc định:

```
admin / password1
user  / password2
```

## API chính

```
POST   /employees
GET    /employees?page=1&limit=20&keyword=backend
GET    /employees/{id}
PUT    /employees/{id}
DELETE /employees/{id}

POST   /employees/export

POST   /departments
GET    /departments?page=1&limit=20
GET    /departments/{id}/employees?page=1&limit=20
```

## Cấu trúc project

```
.
├── cmd/app/main.go
├── internal
│   ├── handlers
│   │   ├── department_handler.go
│   │   └── employee_handler.go
│   ├── middleware
│   │   ├── basic_auth.go
│   │   ├── logging.go
│   │   └── recover.go
│   ├── models
│   │   ├── department.go
│   │   ├── employee.go
│   │   └── pagination.go
│   ├── repositories/postgres
│   │   ├── department_repository.go
│   │   └── employee_repository.go
│   ├── services
│   │   ├── department_service.go
│   │   └── employee_service.go
│   └── utils
│       ├── pagination.go
│       ├── request.go
│       └── response.go
├── migrations
│   ├── 001_create_employees.sql
│   └── 002_create_departments.sql
├── .env.example
├── docker-compose.yaml
├── go.mod
└── README.md
```

Thư mục `exports/` được tạo khi gọi API export và chứa các file export được sinh ra.

## Kiểm tra

```
gofmt -w cmd internal
go test ./...
go test -race ./...
go vet ./...
```
