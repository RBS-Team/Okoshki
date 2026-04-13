package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Конфиг БД по умолчанию для локальной разработки (из .env.dev)
const dsn = "postgres://dev_user:dev_password@localhost:5432/okoshki_db?sslmode=disable"

func main() {
	log.Println("🌱 Запуск сидера базы данных...")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ БД недоступна (проверь, запущен ли docker-compose): %v", err)
	}

	ctx := context.Background()

	// 1. Очищаем старые данные каскадно (чтобы сидер можно было запускать много раз)
	log.Println("🧹 Очистка старых данных...")
	_, err = db.ExecContext(ctx, `TRUNCATE master_services, masters, category, "user" CASCADE;`)
	if err != nil {
		log.Fatalf("❌ Ошибка при очистке таблиц: %v", err)
	}

	// 2. Создаем системных пользователей
	log.Println("👤 Создание пользователей...")
	user1ID := uuid.New()
	user2ID := uuid.New()
	insertUser(ctx, db, user1ID, "anna@okoshki.ru", "Анна (Мастер)")
	insertUser(ctx, db, user2ID, "elena@okoshki.ru", "Елена (Мастер)")

	// 3. Создаем категории
	log.Println("📂 Создание категорий...")
	catNailsID := uuid.New()
	catHairID := uuid.New()
	insertCategory(ctx, db, catNailsID, "Ногти", "Маникюр, педикюр, наращивание")
	insertCategory(ctx, db, catHairID, "Волосы", "Стрижки, укладки, окрашивание")

	// 4. Создаем мастеров
	log.Println("💅 Создание мастеров...")
	master1ID := uuid.New()
	master2ID := uuid.New()

	insertMaster(ctx, db, master1ID, user1ID, "Анна Иванова", "Топ-мастер маникюра с опытом 5 лет. Обожаю френч!", 4.9, 128)
	insertMaster(ctx, db, master2ID, user2ID, "Елена Смирнова", "Стилист по волосам. Делаю идеальный блонд.", 5.0, 45)

	// 5. Создаем услуги (Прайс-лист)
	log.Println("📋 Заполнение прайс-листов...")

	// Услуги Анны (Ногти)
	insertServiceItem(ctx, db, uuid.New(), master1ID, catNailsID, "Маникюр + Гель-лак", "Комбинированный маникюр с однотонным покрытием", 2000, 90, 10, 10)
	insertServiceItem(ctx, db, uuid.New(), master1ID, catNailsID, "Педикюр (Smart)", "Обработка стоп и пальчиков без покрытия", 2500, 60, 10, 10)

	// Услуги Елены (Волосы)
	insertServiceItem(ctx, db, uuid.New(), master2ID, catHairID, "Стрижка женская", "Мытье головы, стрижка, легкая укладка", 1500, 60, 0, 15)
	insertServiceItem(ctx, db, uuid.New(), master2ID, catHairID, "Сложное окрашивание", "Airtouch, Shatush, Balayage", 7000, 240, 15, 30)

	log.Println("✅ База данных успешно заполнена! Фронтенду есть с чем работать.")
}

func insertUser(ctx context.Context, db *sql.DB, id uuid.UUID, email, name string) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO "user" (user_id, email, password_hash, role) 
		VALUES ($1, $2, 'fake_hash_123', 'master')
	`, id, email)
	if err != nil {
		log.Fatalf("Ошибка вставки юзера: %v", err)
	}
}

func insertCategory(ctx context.Context, db *sql.DB, id uuid.UUID, name, desc string) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO category (id, name, description) 
		VALUES ($1, $2, $3)
	`, id, name, desc)
	if err != nil {
		log.Fatalf("Ошибка вставки категории: %v", err)
	}
}

func insertMaster(ctx context.Context, db *sql.DB, id, userID uuid.UUID, name, bio string, rating float64, reviews int) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO masters (id, user_id, name, bio, timezone, lat, lon, rating, review_count) 
		VALUES ($1, $2, $3, $4, 'Europe/Moscow', 55.7558, 37.6173, $5, $6)
	`, id, userID, name, bio, rating, reviews)
	if err != nil {
		log.Fatalf("Ошибка вставки мастера: %v", err)
	}
}

func insertServiceItem(ctx context.Context, db *sql.DB, id, masterID, catID uuid.UUID, title, desc string, price float64, duration, bufBefore, bufAfter int) {
	_, err := db.ExecContext(ctx, `
		INSERT INTO master_services (
			id, master_id, category_id, title, description, 
			price, duration_minutes, buffer_before_minutes, buffer_after_minutes
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, masterID, catID, title, desc, price, duration, bufBefore, bufAfter)
	if err != nil {
		log.Fatalf("Ошибка вставки услуги: %v", err)
	}
}
