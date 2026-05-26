# Tech Stack — Pinned Versions

Authoritative source: `go.mod`. Update this memory when a direct dep's major or load-bearing minor changes.

## Runtime / language

- Go `1.25.0` (per `go.mod`; toolchain-managed). No vendor dir — modules only.
- Postgres `14+` (uses `gen_random_uuid()` from `pgcrypto`; partial unique indexes; `BEFORE UPDATE` triggers).
- Darwin dev host; Linux Docker prod (see `Dockerfile` + `docker-compose.yml`).

## Direct dependencies (from `go.mod` require block)

| Concern | Package | Version |
|---|---|---|
| HTTP framework | `github.com/gin-gonic/gin` | v1.12.0 |
| ORM | `gorm.io/gorm` | v1.31.1 |
| Postgres driver | `gorm.io/driver/postgres` | v1.6.0 |
| Migrations | `github.com/golang-migrate/migrate/v4` | v4.19.1 |
| JWT (HS256) | `github.com/golang-jwt/jwt/v5` | v5.3.1 |
| Bcrypt (cost 12) | `golang.org/x/crypto` | v0.51.0 |
| UUID | `github.com/google/uuid` | v1.6.0 |
| Env loader | `github.com/joho/godotenv` | v1.5.1 |
| AWS SDK v2 (S3) | `github.com/aws/aws-sdk-go-v2` + `service/s3` | v1.41.7 / v1.101.0 |
| Swagger gen | `github.com/swaggo/swag` | v1.16.6 |
| Swagger UI handler | `github.com/swaggo/gin-swagger` + `swaggo/files` | v1.6.1 / v1.0.1 |
| Test assertions | `github.com/stretchr/testify` | v1.11.1 |

## Notable stack choices

- **Storage**: AWS SDK v2 client pointed at Supabase S3-compatible endpoint via `STORAGE_*` env. `NewUploadService` does NOT ping S3 at boot — storage is optional locally.
- **Realtime**: Server-Sent Events hub (`internal/sse/`) — introduced Phase 7; may not exist on `main` yet. Check before importing.
- **Validation**: `go-playground/validator/v10` (transitive via Gin). DTOs are the validation boundary.
- **MIME sniffing for uploads**: `http.DetectContentType` + allowlist; client `Content-Type` is a hint only (review-fix pattern from Phase 2).
- **No SQLite mock for tests**: integration tests hit a real Postgres test DB (`exnodes_hrm_test` by default, or `TEST_DATABASE_URL`).
