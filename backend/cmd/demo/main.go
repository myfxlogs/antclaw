// Demo program to test data collection and storage
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"

	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/infra/redis"
	"github.com/antclaw/antclaw/internal/service/calendar"
)

func main() {
	fmt.Println("=== AntClaw Data Collection Demo ===")
	fmt.Println()

	// Connect to database
	fmt.Println("Connecting to PostgreSQL...")
	dbpool, err := pgxpool.New(context.Background(), "postgres://antclaw:antclaw@localhost:5432/antclaw")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbpool.Close()
	fmt.Println("✓ Database connected")

	// Connect to Redis
	fmt.Println("Connecting to Redis...")
	goredisClient := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
	})
	if err := goredisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("✓ Redis connected")

	// Create our Redis wrapper
	redisClient := redis.NewClient("localhost:6379", "", 0)

	// Create repositories
	calendarRepo := postgres.NewCalendarRepository(dbpool)

	// Create calendar service
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	calendarSvc := calendar.NewCalendarService(calendarRepo, redisClient, logger)

	fmt.Println()
	fmt.Println("=== Testing Data Collection ===")
	fmt.Println()

	// Test 1: Fetch calendar events
	fmt.Println("1. Fetching economic calendar events from MQL5...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := calendarSvc.SyncWeek(ctx)
	if err != nil {
		fmt.Printf("   ⚠ Calendar sync: %v\n", err)
		fmt.Println("   (Note: This may fail if MQL5 endpoint is not accessible)")
	} else {
		fmt.Printf("   ✓ Calendar sync completed: %d events inserted\n", result.Inserted)
	}

	// Test 2: Check database tables
	fmt.Println()
	fmt.Println("2. Checking database tables...")
	checkTables(dbpool)

	// Test 3: Redis cache test
	fmt.Println()
	fmt.Println("3. Testing Redis cache...")
	testRedis(redisClient)

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
}

func checkTables(dbpool *pgxpool.Pool) {
	ctx := context.Background()

	// Query table counts
	tables := []string{
		"calendar_events",
		"cot_records",
		"price_daily",
		"data_snapshots",
	}

	for _, table := range tables {
		var count int64
		err := dbpool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			fmt.Printf("   ✗ %s: %v\n", table, err)
		} else {
			fmt.Printf("   ✓ %s: %d records\n", table, count)
		}
	}
}

func testRedis(client *redis.Client) {
	ctx := context.Background()
	key := "demo:test"
	value := "Hello from AntClaw"

	// Set value
	err := client.Set(ctx, key, value, 60*time.Second)
	if err != nil {
		fmt.Printf("   ✗ Redis set failed: %v\n", err)
		return
	}

	// Get value
	val, err := client.Get(ctx, key)
	if err != nil {
		fmt.Printf("   ✗ Redis get failed: %v\n", err)
		return
	}

	fmt.Printf("   ✓ Redis cache working: %s\n", val)
}
