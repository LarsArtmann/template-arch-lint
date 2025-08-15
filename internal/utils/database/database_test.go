package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/utils/database"
)

func TestDatabase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Utils Suite")
}

var _ = Describe("ConnectionManager", func() {
	var (
		logger *slog.Logger
		cm     *database.ConnectionManager
		ctx    context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		cm = database.NewConnectionManager(logger)
		ctx = context.Background()
	})

	Describe("Connect", func() {
		It("should establish a connection successfully", func() {
			config := database.InMemoryTestConfig()
			db, err := cm.Connect(ctx, config)
			Expect(err).ToNot(HaveOccurred())
			Expect(db).ToNot(BeNil())

			// Clean up
			err = cm.Disconnect(db)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle invalid configuration", func() {
			db, err := cm.Connect(ctx, nil)
			Expect(err).To(HaveOccurred())
			Expect(db).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("configuration cannot be nil"))
		})

		It("should handle invalid driver", func() {
			config := &database.ConnectionConfig{
				Driver: "invalid_driver",
				DSN:    ":memory:",
			}
			db, err := cm.Connect(ctx, config)
			Expect(err).To(HaveOccurred())
			Expect(db).To(BeNil())
		})
	})

	Describe("ConnectWithDefaults", func() {
		It("should connect with default configuration", func() {
			db, err := cm.ConnectWithDefaults(ctx, "sqlite3", ":memory:")
			Expect(err).ToNot(HaveOccurred())
			Expect(db).ToNot(BeNil())

			// Clean up
			err = cm.Disconnect(db)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("ConnectInMemory", func() {
		It("should connect to in-memory database", func() {
			db, err := cm.ConnectInMemory(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(db).ToNot(BeNil())

			// Clean up
			err = cm.Disconnect(db)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("HealthCheck", func() {
		It("should pass for healthy connection", func() {
			db, err := cm.ConnectInMemory(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = cm.HealthCheck(ctx, db)
			Expect(err).ToNot(HaveOccurred())

			// Clean up
			err = cm.Disconnect(db)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail for nil connection", func() {
			err := cm.HealthCheck(ctx, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection is nil"))
		})
	})

	Describe("RunInTransaction", func() {
		It("should commit successful transaction", func() {
			db, err := cm.ConnectInMemory(ctx)
			Expect(err).ToNot(HaveOccurred())
			defer cm.Disconnect(db)

			// Create test table
			_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, value TEXT)")
			Expect(err).ToNot(HaveOccurred())

			err = cm.RunInTransaction(ctx, db, func(tx *sql.Tx) error {
				_, err := tx.Exec("INSERT INTO test_table (value) VALUES (?)", "test_value")
				return err
			})

			Expect(err).ToNot(HaveOccurred())

			// Verify data was committed
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
			Expect(err).ToNot(HaveOccurred())
			Expect(count).To(Equal(1))
		})

		It("should rollback failed transaction", func() {
			db, err := cm.ConnectInMemory(ctx)
			Expect(err).ToNot(HaveOccurred())
			defer cm.Disconnect(db)

			// Create test table
			_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, value TEXT)")
			Expect(err).ToNot(HaveOccurred())

			err = cm.RunInTransaction(ctx, db, func(tx *sql.Tx) error {
				_, err := tx.Exec("INSERT INTO test_table (value) VALUES (?)", "test_value")
				if err != nil {
					return err
				}
				return fmt.Errorf("simulated error")
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("simulated error"))

			// Verify data was rolled back
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
			Expect(err).ToNot(HaveOccurred())
			Expect(count).To(Equal(0))
		})
	})

	Describe("Disconnect", func() {
		It("should close connection successfully", func() {
			db, err := cm.ConnectInMemory(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = cm.Disconnect(db)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle nil connection", func() {
			err := cm.Disconnect(nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("Configuration Functions", func() {
	Describe("DefaultSQLiteConfig", func() {
		It("should return valid configuration", func() {
			config := database.DefaultSQLiteConfig("/path/to/database.db")
			Expect(config).ToNot(BeNil())
			Expect(config.Driver).To(Equal("sqlite3"))
			Expect(config.DSN).To(Equal("/path/to/database.db"))
			Expect(config.MaxOpenConns).To(BeNumerically(">", 0))
			Expect(config.MaxIdleConns).To(BeNumerically(">", 0))
		})
	})

	Describe("InMemoryTestConfig", func() {
		It("should return valid in-memory configuration", func() {
			config := database.InMemoryTestConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Driver).To(Equal("sqlite3"))
			Expect(config.DSN).To(Equal(":memory:"))
			Expect(config.MaxOpenConns).To(Equal(1))
			Expect(config.MaxIdleConns).To(Equal(1))
		})
	})
})

var _ = Describe("Convenience Functions", func() {
	var (
		logger *slog.Logger
		ctx    context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		ctx = context.Background()
	})

	Describe("QuickConnect", func() {
		It("should connect to SQLite database", func() {
			db, err := database.QuickConnect(ctx, ":memory:", logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(db).ToNot(BeNil())

			// Clean up
			database.SafeClose(db, logger)
		})
	})

	Describe("QuickConnectInMemory", func() {
		It("should connect to in-memory database", func() {
			db, err := database.QuickConnectInMemory(ctx, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(db).ToNot(BeNil())

			// Clean up
			database.SafeClose(db, logger)
		})
	})

	Describe("SafeClose", func() {
		It("should handle nil database safely", func() {
			// Should not panic
			database.SafeClose(nil, logger)
		})

		It("should close database without error", func() {
			db, err := database.QuickConnectInMemory(ctx, logger)
			Expect(err).ToNot(HaveOccurred())

			// Should not panic or error
			database.SafeClose(db, logger)
		})
	})
})
