# ===== Build Stage =====
FROM golang:1.22-alpine AS builder

WORKDIR /usr/src/app

# ติดตั้ง dependencies สำหรับ build
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code ทั้งหมด
COPY . .

# Build เป็น binary แบบ static (ไม่ต้องใช้ CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server main.go

# ===== Production Stage =====
FROM alpine:3.20

WORKDIR /usr/src/app

# ติดตั้ง Timezone (Bangkok) และ SSL Certificates
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Bangkok

# สร้าง user ที่ไม่ใช่ root เพื่อความปลอดภัย
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy ไฟล์ binary ที่ build เสร็จแล้วมาจาก builder
COPY --from=builder /usr/src/app/server .

RUN chown -R appuser:appgroup /usr/src/app
USER appuser

EXPOSE 4000

# รัน binary โดยตรง (ไม่ต้องใช้ PM2 เพราะ Go จัดการ thread และ Docker restart ได้เอง)
CMD ["./server"]