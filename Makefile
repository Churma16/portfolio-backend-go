# Ganti 'secret' dengan password asli Anda!
DB_URL=postgresql://postgres:root@localhost:5432/go_portfolio_db?sslmode=disable

# 1. Buat file migrasi baru
# Cara pakai: make new_migration name=init_schema
new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

# 2. Kirim tabel ke Database (UP)
migrate_up:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

# 3. Hapus tabel dari Database (DOWN/Rollback)
migrate_down:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

# 4. Generate kode Go dari SQL
sqlc:
	sqlc generate

# 5. Jalanin Server
run:
	go run main.go

.PHONY: new_migration migrate_up migrate_down sqlc run