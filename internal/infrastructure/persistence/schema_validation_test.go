package persistence_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/internal/infrastructure/persistence"
)

func TestSchemaValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "🗄️ Database Schema Validation Suite - Domain Model Consistency")
}

// Schema validation helpers
type ColumnInfo struct {
	Name         string
	Type         string
	NotNull      bool
	DefaultValue sql.NullString
	PrimaryKey   bool
}

type TableInfo struct {
	Name    string
	Columns []ColumnInfo
	Indexes []IndexInfo
}

type IndexInfo struct {
	Name    string
	Unique  bool
	Columns []string
}

var _ = Describe("🗄️ Database Schema Validation Against Domain Model", func() {
	var (
		db     *sql.DB
		ctx    context.Context
		logger *slog.Logger
		repo   *persistence.SQLCUserRepository
	)

	setupInMemoryDB := func() *sql.DB {
		database, err := sql.Open("sqlite3", ":memory:")
		Expect(err).ToNot(HaveOccurred())
		return database
	}

	cleanupDB := func(database *sql.DB) {
		if database != nil {
			err := database.Close()
			Expect(err).ToNot(HaveOccurred())
		}
	}

	getTableInfo := func(tableName string) TableInfo {
		// Get column information
		rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
		Expect(err).ToNot(HaveOccurred())
		defer rows.Close()

		var columns []ColumnInfo
		for rows.Next() {
			var cid int
			var col ColumnInfo
			var dfltValue sql.NullString
			var pk int

			err := rows.Scan(&cid, &col.Name, &col.Type, &col.NotNull, &dfltValue, &pk)
			Expect(err).ToNot(HaveOccurred())

			col.DefaultValue = dfltValue
			col.PrimaryKey = pk == 1
			columns = append(columns, col)
		}

		// Get index information
		indexRows, err := db.Query("PRAGMA index_list(" + tableName + ")")
		Expect(err).ToNot(HaveOccurred())
		defer indexRows.Close()

		var indexes []IndexInfo
		for indexRows.Next() {
			var seq int
			var index IndexInfo
			var origin string

			err := indexRows.Scan(&seq, &index.Name, &index.Unique, &origin)
			Expect(err).ToNot(HaveOccurred())

			// Get columns for this index
			colRows, err := db.Query("PRAGMA index_info(" + index.Name + ")")
			Expect(err).ToNot(HaveOccurred())

			for colRows.Next() {
				var seqno, cid int
				var name string
				err := colRows.Scan(&seqno, &cid, &name)
				Expect(err).ToNot(HaveOccurred())
				index.Columns = append(index.Columns, name)
			}
			colRows.Close()

			indexes = append(indexes, index)
		}

		return TableInfo{
			Name:    tableName,
			Columns: columns,
			Indexes: indexes,
		}
	}

	BeforeEach(func() {
		ctx = context.Background()
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		db = setupInMemoryDB()
		repo = persistence.NewSQLCUserRepository(db, logger)
	})

	AfterEach(func() {
		cleanupDB(db)
	})

	Describe("📋 Schema Structure Validation", func() {
		Context("users table schema", func() {
			It("should have the correct table structure for User entity", func() {
				tableInfo := getTableInfo("users")

				// Verify table exists
				Expect(tableInfo.Name).To(Equal("users"))
				Expect(tableInfo.Columns).ToNot(BeEmpty())

				// Expected columns based on User entity
				expectedColumns := map[string]struct {
					Type       string
					NotNull    bool
					PrimaryKey bool
				}{
					"id":       {"TEXT", true, true},
					"email":    {"TEXT", true, false},
					"name":     {"TEXT", true, false},
					"created":  {"DATETIME", true, false},
					"modified": {"DATETIME", true, false},
				}

				// Verify all expected columns exist
				actualColumns := make(map[string]ColumnInfo)
				for _, col := range tableInfo.Columns {
					actualColumns[col.Name] = col
				}

				for expectedName, expectedProps := range expectedColumns {
					actualCol, exists := actualColumns[expectedName]
					Expect(exists).To(BeTrue(), "Column %s should exist", expectedName)

					Expect(strings.ToUpper(actualCol.Type)).To(ContainSubstring(expectedProps.Type),
						"Column %s should have type %s, got %s", expectedName, expectedProps.Type, actualCol.Type)

					Expect(actualCol.NotNull).To(Equal(expectedProps.NotNull),
						"Column %s NotNull should be %t, got %t", expectedName, expectedProps.NotNull, actualCol.NotNull)

					Expect(actualCol.PrimaryKey).To(Equal(expectedProps.PrimaryKey),
						"Column %s PrimaryKey should be %t, got %t", expectedName, expectedProps.PrimaryKey, actualCol.PrimaryKey)
				}

				// Verify no unexpected columns
				for actualName := range actualColumns {
					_, expected := expectedColumns[actualName]
					Expect(expected).To(BeTrue(), "Unexpected column: %s", actualName)
				}
			})

			It("should have appropriate constraints and indexes", func() {
				tableInfo := getTableInfo("users")

				// Verify email uniqueness constraint exists
				emailIndexExists := false
				for _, index := range tableInfo.Indexes {
					if index.Unique {
						for _, col := range index.Columns {
							if col == "email" {
								emailIndexExists = true
								break
							}
						}
					}
				}

				Expect(emailIndexExists).To(BeTrue(), "Should have unique constraint on email column")

				// Verify primary key constraint
				primaryKeyExists := false
				for _, col := range tableInfo.Columns {
					if col.Name == "id" && col.PrimaryKey {
						primaryKeyExists = true
						break
					}
				}

				Expect(primaryKeyExists).To(BeTrue(), "Should have primary key on id column")
			})
		})
	})

	Describe("🔄 Domain Entity Mapping Validation", func() {
		Context("User entity field mapping", func() {
			It("should correctly map all User entity fields to database columns", func() {
				// Create a test user through the entity
				userID, err := values.NewUserID("schema-test-user")
				Expect(err).ToNot(HaveOccurred())

				user, err := entities.NewUser(userID, "schema@example.com", "Schema Test User")
				Expect(err).ToNot(HaveOccurred())

				// Save through repository
				err = repo.Save(ctx, user)
				Expect(err).ToNot(HaveOccurred())

				// Verify data mapping by querying directly
				var dbID, dbEmail, dbName string
				var dbCreated, dbModified time.Time

				query := "SELECT id, email, name, created, modified FROM users WHERE id = ?"
				err = db.QueryRowContext(ctx, query, userID.String()).Scan(
					&dbID, &dbEmail, &dbName, &dbCreated, &dbModified)
				Expect(err).ToNot(HaveOccurred())

				// Verify correct mapping
				Expect(dbID).To(Equal(user.ID.String()))
				Expect(dbEmail).To(Equal(user.Email))
				Expect(dbName).To(Equal(user.Name))
				Expect(dbCreated).To(BeTemporally("~", user.Created, time.Second))
				Expect(dbModified).To(BeTemporally("~", user.Modified, time.Second))
			})

			It("should handle all domain value object types correctly", func() {
				// Test various UserID formats
				userIDFormats := []string{
					"simple-id",
					"uuid-like-550e8400-e29b-41d4-a716-446655440000",
					"user_with_underscores",
					"user-with-hyphens",
					"MixedCaseUserID123",
				}

				for i, idStr := range userIDFormats {
					userID, err := values.NewUserID(idStr)
					Expect(err).ToNot(HaveOccurred())

					email := fmt.Sprintf("test%d@example.com", i)
					name := fmt.Sprintf("Test User %d", i)

					user, err := entities.NewUser(userID, email, name)
					Expect(err).ToNot(HaveOccurred())

					err = repo.Save(ctx, user)
					Expect(err).ToNot(HaveOccurred())

					// Verify retrieval
					retrievedUser, err := repo.FindByID(ctx, userID)
					Expect(err).ToNot(HaveOccurred())
					Expect(retrievedUser.ID.String()).To(Equal(idStr))
					Expect(retrievedUser.Email).To(Equal(email))
					Expect(retrievedUser.Name).To(Equal(name))
				}
			})

			It("should handle time field precision correctly", func() {
				userID, err := values.NewUserID("time-test-user")
				Expect(err).ToNot(HaveOccurred())

				user, err := entities.NewUser(userID, "time@example.com", "Time Test User")
				Expect(err).ToNot(HaveOccurred())

				originalCreated := user.Created
				originalModified := user.Modified

				err = repo.Save(ctx, user)
				Expect(err).ToNot(HaveOccurred())

				// Retrieve and check time precision
				retrievedUser, err := repo.FindByID(ctx, userID)
				Expect(err).ToNot(HaveOccurred())

				// Times should be preserved with reasonable precision (within 1 second)
				Expect(retrievedUser.Created).To(BeTemporally("~", originalCreated, time.Second))
				Expect(retrievedUser.Modified).To(BeTemporally("~", originalModified, time.Second))

				// Update user and verify modified time changes
				time.Sleep(10 * time.Millisecond) // Ensure time difference
				err = retrievedUser.SetEmail("updated@example.com")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Update(ctx, retrievedUser)
				Expect(err).ToNot(HaveOccurred())

				updatedUser, err := repo.FindByID(ctx, userID)
				Expect(err).ToNot(HaveOccurred())

				// Created should remain the same, modified should be updated
				Expect(updatedUser.Created).To(BeTemporally("~", originalCreated, time.Second))
				Expect(updatedUser.Modified).To(BeTemporally(">", originalModified))
			})
		})
	})

	Describe("🛡️ Data Integrity Validation", func() {
		Context("constraint enforcement", func() {
			It("should enforce primary key constraint", func() {
				userID, err := values.NewUserID("pk-test-user")
				Expect(err).ToNot(HaveOccurred())

				// Create first user
				user1, err := entities.NewUser(userID, "pk1@example.com", "PK Test User 1")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Save(ctx, user1)
				Expect(err).ToNot(HaveOccurred())

				// Try to create second user with same ID (should fail)
				user2, err := entities.NewUser(userID, "pk2@example.com", "PK Test User 2")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Save(ctx, user2)
				Expect(err).To(HaveOccurred(), "Should fail due to primary key constraint")
			})

			It("should enforce email uniqueness constraint", func() {
				email := "unique@example.com"

				// Create first user
				userID1, err := values.NewUserID("unique-user-1")
				Expect(err).ToNot(HaveOccurred())

				user1, err := entities.NewUser(userID1, email, "Unique User 1")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Save(ctx, user1)
				Expect(err).ToNot(HaveOccurred())

				// Try to create second user with same email (should fail)
				userID2, err := values.NewUserID("unique-user-2")
				Expect(err).ToNot(HaveOccurred())

				user2, err := entities.NewUser(userID2, email, "Unique User 2")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Save(ctx, user2)
				Expect(err).To(HaveOccurred(), "Should fail due to email uniqueness constraint")
			})

			It("should enforce NOT NULL constraints", func() {
				// Test with direct SQL to bypass domain validation
				queries := []string{
					"INSERT INTO users (id, email, name, created, modified) VALUES (NULL, 'test@example.com', 'Test', datetime('now'), datetime('now'))",
					"INSERT INTO users (id, email, name, created, modified) VALUES ('test-id', NULL, 'Test', datetime('now'), datetime('now'))",
					"INSERT INTO users (id, email, name, created, modified) VALUES ('test-id', 'test@example.com', NULL, datetime('now'), datetime('now'))",
					"INSERT INTO users (id, email, name, created, modified) VALUES ('test-id', 'test@example.com', 'Test', NULL, datetime('now'))",
					"INSERT INTO users (id, email, name, created, modified) VALUES ('test-id', 'test@example.com', 'Test', datetime('now'), NULL)",
				}

				for _, query := range queries {
					_, err := db.ExecContext(ctx, query)
					Expect(err).To(HaveOccurred(), "Should fail due to NOT NULL constraint: %s", query)
				}
			})
		})
	})

	Describe("🔍 Schema Evolution Validation", func() {
		Context("schema compatibility", func() {
			It("should handle domain model extensions gracefully", func() {
				// This test ensures that if we add new fields to the User entity,
				// the existing schema can still handle current operations

				// Current schema should work with current operations
				userID, err := values.NewUserID("evolution-test-user")
				Expect(err).ToNot(HaveOccurred())

				user, err := entities.NewUser(userID, "evolution@example.com", "Evolution Test User")
				Expect(err).ToNot(HaveOccurred())

				// Full CRUD cycle
				err = repo.Save(ctx, user)
				Expect(err).ToNot(HaveOccurred())

				retrievedUser, err := repo.FindByID(ctx, userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(retrievedUser).ToNot(BeNil())

				err = retrievedUser.SetEmail("updated-evolution@example.com")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Update(ctx, retrievedUser)
				Expect(err).ToNot(HaveOccurred())

				err = repo.Delete(ctx, userID)
				Expect(err).ToNot(HaveOccurred())

				// All operations should complete successfully
			})

			It("should verify schema supports all repository operations", func() {
				// Verify schema supports all methods defined in UserRepository interface
				repositoryInterface := reflect.TypeOf((*repositories.UserRepository)(nil)).Elem()

				// Get all methods from the interface
				methods := make([]string, repositoryInterface.NumMethod())
				for i := 0; i < repositoryInterface.NumMethod(); i++ {
					methods[i] = repositoryInterface.Method(i).Name
				}

				// Verify each method can be executed (basic smoke test)
				userID, err := values.NewUserID("interface-test-user")
				Expect(err).ToNot(HaveOccurred())

				user, err := entities.NewUser(userID, "interface@example.com", "Interface Test User")
				Expect(err).ToNot(HaveOccurred())

				// Test each repository method
				for _, methodName := range methods {
					By(fmt.Sprintf("Testing repository method: %s", methodName))

					switch methodName {
					case "Save":
						err = repo.Save(ctx, user)
						Expect(err).ToNot(HaveOccurred())

					case "FindByID":
						foundUser, err := repo.FindByID(ctx, userID)
						Expect(err).ToNot(HaveOccurred())
						Expect(foundUser).ToNot(BeNil())

					case "FindByEmail":
						foundUser, err := repo.FindByEmail(ctx, "interface@example.com")
						Expect(err).ToNot(HaveOccurred())
						Expect(foundUser).ToNot(BeNil())

					case "FindByUsername":
						foundUser, err := repo.FindByUsername(ctx, "Interface Test User")
						Expect(err).ToNot(HaveOccurred())
						Expect(foundUser).ToNot(BeNil())

					case "Update":
						err = user.SetName("Updated Interface Test User")
						Expect(err).ToNot(HaveOccurred())
						err = repo.Update(ctx, user)
						Expect(err).ToNot(HaveOccurred())

					case "List":
						users, err := repo.List(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(users).ToNot(BeNil())

					case "Count":
						count, err := repo.Count(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(count).To(BeNumerically(">=", 0))

					case "Exists":
						exists, err := repo.Exists(ctx, userID)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(BeTrue())

					case "Delete":
						// Test delete last to clean up
						err = repo.Delete(ctx, userID)
						Expect(err).ToNot(HaveOccurred())
					}
				}
			})
		})
	})

	Describe("🎯 Performance Schema Validation", func() {
		Context("query performance", func() {
			It("should have efficient indexes for common queries", func() {
				// Test that common query patterns are efficient
				// We'll measure this by checking that queries complete quickly

				const numUsers = 100
				userIDs := make([]values.UserID, numUsers)

				// Create test data
				for i := 0; i < numUsers; i++ {
					userID, err := values.NewUserID(fmt.Sprintf("perf-user-%d", i))
					Expect(err).ToNot(HaveOccurred())

					user, err := entities.NewUser(userID, fmt.Sprintf("perf%d@example.com", i), fmt.Sprintf("Perf User %d", i))
					Expect(err).ToNot(HaveOccurred())

					err = repo.Save(ctx, user)
					Expect(err).ToNot(HaveOccurred())

					userIDs[i] = userID
				}

				// Test FindByID performance (should use primary key index)
				start := time.Now()
				for i := 0; i < 10; i++ {
					_, err := repo.FindByID(ctx, userIDs[i])
					Expect(err).ToNot(HaveOccurred())
				}
				findByIDDuration := time.Since(start)

				// Test FindByEmail performance (should use email index)
				start = time.Now()
				for i := 0; i < 10; i++ {
					_, err := repo.FindByEmail(ctx, fmt.Sprintf("perf%d@example.com", i))
					Expect(err).ToNot(HaveOccurred())
				}
				findByEmailDuration := time.Since(start)

				// Performance should be reasonable
				Expect(findByIDDuration).To(BeNumerically("<", 100*time.Millisecond), "FindByID should be fast")
				Expect(findByEmailDuration).To(BeNumerically("<", 100*time.Millisecond), "FindByEmail should be fast")

				By(fmt.Sprintf("FindByID performance: %v for 10 lookups", findByIDDuration))
				By(fmt.Sprintf("FindByEmail performance: %v for 10 lookups", findByEmailDuration))
			})
		})
	})

	Describe("🔬 Schema Validation Edge Cases", func() {
		Context("boundary value testing", func() {
			It("should handle maximum field lengths correctly", func() {
				// Test with maximum reasonable values
				longID := strings.Repeat("a", 100)
				longEmail := strings.Repeat("a", 60) + "@" + strings.Repeat("b", 50) + ".com"
				longName := strings.Repeat("Test User ", 10) // ~100 characters

				userID, err := values.NewUserID(longID)
				Expect(err).ToNot(HaveOccurred())

				user, err := entities.NewUser(userID, longEmail, longName)
				Expect(err).ToNot(HaveOccurred())

				// Should be able to save and retrieve
				err = repo.Save(ctx, user)
				Expect(err).ToNot(HaveOccurred())

				retrievedUser, err := repo.FindByID(ctx, userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(retrievedUser.ID.String()).To(Equal(longID))
				Expect(retrievedUser.Email).To(Equal(longEmail))
				Expect(retrievedUser.Name).To(Equal(longName))
			})

			It("should handle special characters correctly", func() {
				specialCases := []struct {
					id    string
					email string
					name  string
					desc  string
				}{
					{"user-with-hyphens", "test@sub.example.com", "Mary-Jane O'Connor", "hyphens and apostrophes"},
					{"user_with_underscores", "user+tag@example.com", "José María", "underscores and accents"},
					{"UserWithMixedCase123", "MixedCase@Example.COM", "Dr. John Doe Jr.", "mixed case and titles"},
				}

				for _, tc := range specialCases {
					userID, err := values.NewUserID(tc.id)
					Expect(err).ToNot(HaveOccurred())

					user, err := entities.NewUser(userID, tc.email, tc.name)
					Expect(err).ToNot(HaveOccurred())

					err = repo.Save(ctx, user)
					Expect(err).ToNot(HaveOccurred(), "Should handle: %s", tc.desc)

					retrievedUser, err := repo.FindByID(ctx, userID)
					Expect(err).ToNot(HaveOccurred())
					Expect(retrievedUser.ID.String()).To(Equal(tc.id))
					Expect(retrievedUser.Email).To(Equal(tc.email))
					Expect(retrievedUser.Name).To(Equal(tc.name))
				}
			})
		})
	})
})
