# Subglutee Project - Scheduler

cronjob รายวัน 3 ตัว: ล้าง refresh token ที่ใช้ไม่ได้แล้ว, เลื่อนวันตัดเงินไปรอบถัดไป, และ publish reminder เข้าคิวให้ sgt-noti-service ส่งเมล

- ไม่มี business REST API - มีแค่ `GET /health`
- gRPC client -> subscription-service (`ListDueReminders`, `AdvanceBillingDates`)
- gRPC client -> user-service (`GetUser` เพื่อหา email)
- gRPC client -> auth-service (`DeleteExpiredRefreshTokens`)
- RabbitMQ - publish queue `email_notifications` (contract เดียวกับที่ sgt-noti-service consume)
- ไม่มี DB - logic วันที่ทั้งหมดอยู่ที่ subscription-service, scheduler แค่บอก "วันนี้" (YYYY-MM-DD เวลาไทย)

### Jobs

ทุก cron ตีความเป็นเวลา `Asia/Bangkok`

| Job                    | Env             | Default     | ทำอะไร                                                                                                             |
| ---------------------- | --------------- | ----------- | ------------------------------------------------------------------------------------------------------------------ |
| `cleanupExpiredTokens` | `CLEANUP_CRON`  | `0 0 * * *` | ลบ refresh token ที่ revoked หรือหมดอายุใน auth-service                                                            |
| `advanceBillingDates`  | `ROLLOVER_CRON` | `0 0 * * *` | free trial ที่หมดแล้ว -> `active`, `next_billing_date` ที่เลยแล้ว -> รอบถัดไป (ปัดวันสิ้นเดือนตาม `billing_day`)   |
| `checkBillingReminder` | `REMINDER_CRON` | `0 9 * * *` | หา subscription ที่ "วันตัดเงิน / วันหมด trial − `reminder_time_in_advanced`" = วันนี้ แล้ว publish 1 เมลต่อ event |

- `ROLLOVER_CRON` ต้องรันก่อน `REMINDER_CRON` ของวันเดียวกัน ไม่งั้น reminder จะเห็นวันที่เก่า
- ทุก job idempotent: รันซ้ำในวันเดียวกันไม่เลื่อนวันซ้ำ / ไม่ลบซ้ำ ส่วน reminder ซ้ำ noti จะ dedupe ด้วย `reminder_id`
- job ล้ม -> log แล้วรอรอบถัดไป (rollover ตามทันเองเพราะเลื่อนจนถึงวันนี้)

### Message contract (publish)

```json
{
  "reminder_id": "<subscription_id>:billing:2026-10-06",
  "user_id": "11111111-1111-4111-8111-111111111111",
  "subscription_id": "<subscription_id>",
  "to": "user@example.com",
  "title": "Upcoming charge: Netflix",
  "content": "Netflix is expected to charge 419.00 THB on 2026-10-06.\n\nReview it: http://localhost:8000/home"
}
```

| Field                      | Meaning                                                                                      |
| -------------------------- | -------------------------------------------------------------------------------------------- |
| `reminder_id`              | `<subscription_id>:<kind>:<event date>` ค่าเดิมเสมอสำหรับ event เดียวกัน (idempotency key)   |
| `kind` (ใน id)             | `billing` \| `trial_end` \| `trial_end_and_billing` (trial หมดวันเดียวกับตัดเงิน = เมลเดียว) |
| `to` / `title` / `content` | ที่ noti ใช้ส่งเมล - ชื่อ subscription ถูกตัด CR/LF ก่อนใส่ title                            |

- message เป็น persistent (`delivery_mode: 2`) และ `message_id` = `reminder_id`
- publish แบบ publisher confirms: นับว่าส่งแล้วเมื่อ RabbitMQ ack เท่านั้น
- connection RabbitMQ หลุด -> process exit ให้ container restart ใหม่

### Structure

```
scheduler/      cron wrapper (robfig/cron) - Asia/Bangkok, recover panic, skip ถ้ารอบก่อนยังไม่จบ, timeout ต่อรอบ
usecases/       BillingReminder, BillingRollover, TokenCleanup - กำหนด interface ที่ต้องใช้เอง
clients/        gRPC client -> subscription-service, user-service, auth-service (proto adapter)
publisher/      RabbitMQ adapter
routes/         GET /health
config/         env loader
```

flow: `cron tick -> usecases -> clients (gRPC) / publisher (RabbitMQ)`

### Prerequisite

- Go 1.26
- Docker
- `make` - `winget install ezwinports.make`
- tools:

```terminal
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install github.com/evilmartians/lefthook@latest
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Setup

ต้องมี subscription-service (gRPC `:50051`), user-service (gRPC `:50052`), auth-service (gRPC `:50053`) และ RabbitMQ (ของ sgt-noti-service compose) รันอยู่ก่อน

```terminal
git clone https://github.com/polar-bear-cu/sgt-scheduler.git
cd sgt-scheduler
lefthook install
cp .env.example .env
go mod download
make run
```

### Test locally

ตั้ง cron ที่จะลองเป็น `@every 1m` ใน `.env` แล้ว `make run` - log 1 บรรทัดต่อรอบ:

```text
refresh token cleanup: deleted=2
billing rollover: date=2026-10-03 advanced=1 trials_converted=0
reminder check: date=2026-10-03 due=3 published=3 failed=0 took=33ms
```

ดูผล: MailHog `localhost:8025`, RabbitMQ `localhost:15672`, pgweb ของแต่ละ service

ยิง RPC ที่ scheduler ใช้ตรงๆ (cmd):

```terminal
grpcurl -plaintext -d "{\"date\":\"2026-10-03\"}" localhost:50051 subscription.v1.SubscriptionService/ListDueReminders
grpcurl -plaintext -d "{\"date\":\"2026-10-03\"}" localhost:50051 subscription.v1.SubscriptionService/AdvanceBillingDates
grpcurl -plaintext localhost:50053 auth.v1.AuthService/DeleteExpiredRefreshTokens
```

`AdvanceBillingDates` แก้ข้อมูลจริงใน DB - ใช้กับข้อมูลทดสอบเท่านั้น

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
