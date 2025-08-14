package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

// setupBenchmarkRepository creates an in-memory SQLite database for benchmarks.
func setupBenchmarkRepository(b *testing.B) (*SQLUserRepository, context.Context, func()) {
	b.Helper()

	// Create in-memory database for benchmarks
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce logging noise in benchmarks
	}))

	// Create repository
	repo := NewSQLUserRepository(db, logger)

	// Create table schema
	ctx := context.Background()
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
	`

	_, err = db.ExecContext(ctx, createTableSQL)
	if err != nil {
		b.Fatalf("Failed to create table schema: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return repo, ctx, cleanup
}

// populateDatabase fills the database with test data for realistic benchmarks.
func populateDatabase(b *testing.B, repo *SQLUserRepository, ctx context.Context, userCount int) {
	b.Helper()

	for i := 0; i < userCount; i++ {
		userID, err := values.NewUserID(fmt.Sprintf("bench-user-%d", i))
		if err != nil {
			b.Fatalf("Failed to create user ID: %v", err)
		}

		email := fmt.Sprintf("benchuser%d@example.com", i)
		name := fmt.Sprintf("Bench User %d", i)

		user, err := entities.NewUser(userID, email, name)
		if err != nil {
			b.Fatalf("Failed to create user entity: %v", err)
		}

		err = repo.Save(ctx, user)
		if err != nil {
			b.Fatalf("Failed to save user: %v", err)
		}
	}
}

// BenchmarkSQLUserRepository_Save measures user creation performance in SQL.
func BenchmarkSQLUserRepository_Save(b *testing.B) {
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("save-bench-%d", i))
		email := fmt.Sprintf("savebench%d@example.com", i)
		name := fmt.Sprintf("Save Bench User %d", i)

		user, err := entities.NewUser(userID, email, name)
		if err != nil {
			b.Fatalf("Failed to create user: %v", err)
		}

		err = repo.Save(ctx, user)
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

// BenchmarkSQLUserRepository_FindByID measures user retrieval by ID performance.
func BenchmarkSQLUserRepository_FindByID(b *testing.B) {
	const userCount = 10000
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	// Pre-populate database
	populateDatabase(b, repo, ctx, userCount)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i%userCount))

		_, err := repo.FindByID(ctx, userID)
		if err != nil {
			b.Fatalf("FindByID failed: %v", err)
		}
	}
}

// BenchmarkSQLUserRepository_FindByEmail measures user retrieval by email performance.
func BenchmarkSQLUserRepository_FindByEmail(b *testing.B) {
	const userCount = 10000
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	// Pre-populate database
	populateDatabase(b, repo, ctx, userCount)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		email := fmt.Sprintf("benchuser%d@example.com", i%userCount)

		_, err := repo.FindByEmail(ctx, email)
		if err != nil {
			b.Fatalf("FindByEmail failed: %v", err)
		}
	}
}

// BenchmarkSQLUserRepository_List measures user listing performance with different dataset sizes.
func BenchmarkSQLUserRepository_List(b *testing.B) {
	testCases := []struct {
		name      string
		userCount int
	}{
		{"100Users", 100},
		{"1000Users", 1000},
		{"10000Users", 10000},
		{"50000Users", 50000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			repo, ctx, cleanup := setupBenchmarkRepository(b)
			defer cleanup()

			// Pre-populate database
			populateDatabase(b, repo, ctx, tc.userCount)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				users, err := repo.List(ctx)
				if err != nil {
					b.Fatalf("List failed: %v", err)
				}

				// Access the slice to ensure it's not optimized away
				_ = len(users)
			}
		})
	}
}

// BenchmarkSQLUserRepository_Update measures user update performance.
func BenchmarkSQLUserRepository_Update(b *testing.B) {
	const userCount = 1000
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	// Pre-populate database
	populateDatabase(b, repo, ctx, userCount)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i%userCount))

		// Get existing user
		user, err := repo.FindByID(ctx, userID)
		if err != nil {
			b.Fatalf("FindByID failed: %v", err)
		}

		// Update user data
		user.Email = fmt.Sprintf("updated%d@example.com", i)
		user.Name = fmt.Sprintf("Updated User %d", i)

		err = repo.Save(ctx, user)
		if err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

// BenchmarkSQLUserRepository_Delete measures user deletion performance.
func BenchmarkSQLUserRepository_Delete(b *testing.B) {
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	// Pre-populate database with enough users for all benchmark iterations
	populateDatabase(b, repo, ctx, b.N)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i))

		err := repo.Delete(ctx, userID)
		if err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

// BenchmarkSQLUserRepository_ConcurrentReads measures concurrent read performance.
func BenchmarkSQLUserRepository_ConcurrentReads(b *testing.B) {
	const userCount = 10000
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	// Pre-populate database
	populateDatabase(b, repo, ctx, userCount)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			userID, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i%userCount))

			_, err := repo.FindByID(ctx, userID)
			if err != nil {
				b.Fatalf("Concurrent FindByID failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkSQLUserRepository_ConcurrentWrites measures concurrent write performance.
func BenchmarkSQLUserRepository_ConcurrentWrites(b *testing.B) {
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			userID, _ := values.NewUserID(fmt.Sprintf("concurrent-user-%d-%d", b.N, i))
			email := fmt.Sprintf("concurrentuser%d-%d@example.com", b.N, i)
			name := fmt.Sprintf("Concurrent User %d-%d", b.N, i)

			user, err := entities.NewUser(userID, email, name)
			if err != nil {
				b.Fatalf("Failed to create user: %v", err)
			}

			err = repo.Save(ctx, user)
			if err != nil {
				b.Fatalf("Concurrent Save failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkSQLUserRepository_MixedOperations measures realistic mixed workload.
func BenchmarkSQLUserRepository_MixedOperations(b *testing.B) {
	const userCount = 1000
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	// Pre-populate database
	populateDatabase(b, repo, ctx, userCount)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		switch i % 10 {
		case 0, 1, 2, 3, 4, 5: // 60% reads (most common operation)
			userID, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i%userCount))
			_, _ = repo.FindByID(ctx, userID)
		case 6, 7: // 20% email lookups
			email := fmt.Sprintf("benchuser%d@example.com", i%userCount)
			_, _ = repo.FindByEmail(ctx, email)
		case 8: // 10% updates
			userID, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i%userCount))
			if user, err := repo.FindByID(ctx, userID); err == nil {
				user.Name = fmt.Sprintf("Mixed Updated User %d", i)
				_ = repo.Save(ctx, user)
			}
		case 9: // 10% new inserts
			userID, _ := values.NewUserID(fmt.Sprintf("mixed-new-user-%d", i))
			email := fmt.Sprintf("mixednew%d@example.com", i)
			name := fmt.Sprintf("Mixed New User %d", i)
			if user, err := entities.NewUser(userID, email, name); err == nil {
				_ = repo.Save(ctx, user)
			}
		}
	}
}

// BenchmarkSQLUserRepository_BatchOperations measures batch insert performance.
func BenchmarkSQLUserRepository_BatchOperations(b *testing.B) {
	batchSizes := []int{10, 50, 100, 500}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize%d", batchSize), func(b *testing.B) {
			repo, ctx, cleanup := setupBenchmarkRepository(b)
			defer cleanup()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Simulate batch insert by doing multiple saves in sequence
				for j := 0; j < batchSize; j++ {
					userID, _ := values.NewUserID(fmt.Sprintf("batch-%d-%d", i, j))
					email := fmt.Sprintf("batch%d-%d@example.com", i, j)
					name := fmt.Sprintf("Batch User %d-%d", i, j)

					user, err := entities.NewUser(userID, email, name)
					if err != nil {
						b.Fatalf("Failed to create user: %v", err)
					}

					err = repo.Save(ctx, user)
					if err != nil {
						b.Fatalf("Batch save failed: %v", err)
					}
				}
			}
		})
	}
}

// BenchmarkSQLUserRepository_MemoryUsage measures memory allocation patterns.
func BenchmarkSQLUserRepository_MemoryUsage(b *testing.B) {
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create and save user (allocates structs, strings, etc.)
		userID, _ := values.NewUserID(fmt.Sprintf("memory-test-%d", i))
		email := fmt.Sprintf("memorytest%d@example.com", i)
		name := fmt.Sprintf("Memory Test User %d", i)

		user, err := entities.NewUser(userID, email, name)
		if err != nil {
			b.Fatalf("Failed to create user: %v", err)
		}

		err = repo.Save(ctx, user)
		if err != nil {
			b.Fatalf("Save failed: %v", err)
		}

		// Immediately read it back (allocates during scan)
		_, err = repo.FindByID(ctx, userID)
		if err != nil {
			b.Fatalf("FindByID failed: %v", err)
		}
	}
}

// BenchmarkSQLUserRepository_QueryComplexity measures different query patterns.
func BenchmarkSQLUserRepository_QueryComplexity(b *testing.B) {
	const userCount = 10000
	repo, ctx, cleanup := setupBenchmarkRepository(b)
	defer cleanup()

	// Pre-populate database
	populateDatabase(b, repo, ctx, userCount)

	b.Run("SimpleIDLookup", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			userID, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i%userCount))
			_, _ = repo.FindByID(ctx, userID)
		}
	})

	b.Run("IndexedEmailLookup", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			email := fmt.Sprintf("benchuser%d@example.com", i%userCount)
			_, _ = repo.FindByEmail(ctx, email)
		}
	})

	b.Run("FullTableScan", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = repo.List(ctx)
		}
	})
}
