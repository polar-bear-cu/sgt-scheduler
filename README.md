# Subglutee Project - Scheduler

cronjob สำหรับหา subscription ที่ใกล้ billing แล้ว publish reminder เข้าคิวให้ sgt-noti-service ส่งเมล

- ไม่มี business REST API - มีแค่ `GET /health`
- gRPC client -> subscription-service (`GetUpcomingForBilling`)
- gRPC client -> user-service (`GetUser`)
- RabbitMQ - publish queue `email_notifications` (contract เดียวกับที่ sgt-noti-service consume)
- ไม่มี DB

### Message contract (publish)

```json
{ "to": "user@example.com", "subject": "...", "body": "..." }
```

### Structure

```
scheduler/      cron wrapper (robfig/cron)
usecases/       BillingReminderUsecase - ผูก subscription source + user directory + publisher
clients/        gRPC client -> subscription-service, user-service (proto adapter)
publisher/      RabbitMQ adapter (usecase กำหนด interface เอง)
routes/         GET /health
config/         env loader
```

flow: `cron tick -> usecases -> clients (หา subscription + email) -> publisher (ส่งเข้าคิว)`

### Prerequisite

- Go 1.26
- Docker
- `make` - `winget install ezwinports.make`
- tools:

```terminal
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install github.com/evilmartians/lefthook@latest
```

### Setup

ต้องมี subscription-service (gRPC `:50051`), user-service (gRPC `:50052`) และ RabbitMQ (ของ sgt-noti-service compose) รันอยู่ก่อน

```terminal
git clone https://github.com/polar-bear-cu/sgt-scheduler.git
cd sgt-scheduler
lefthook install
cp .env.example .env
go mod download
make run
```

### Run alternatively (container)

```terminal
make image
make container
```

### Useful Commands

Check `Makefile`

#### ถ้าแก้ proto พร้อม service นี้

```terminal
cd ..
go work use ./sgt-proto ./sgt-scheduler
```
