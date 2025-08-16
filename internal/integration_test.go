package internal_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/internal/infrastructure/persistence"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "🔗 Integration Testing Suite - Cross-Layer Component Testing")
}

var _ = Describe("🔗 Integration Testing - Cross-Layer Component Interaction", func() {
	var (
		ctx         context.Context
		db          *sql.DB
		logger      *slog.Logger
		userRepo    repositories.UserRepository
		userService *services.UserService
		_           *handlers.UserHandler // userHandler for potential future use
	)

	// Test helper functions
	createTestUserID := func(id string) values.UserID {
		userID, err := values.NewUserID(id)
		Expect(err).ToNot(HaveOccurred())
		return userID
	}

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

	BeforeEach(func() {
		ctx = context.Background()
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		db = setupInMemoryDB()

		// Build the full component stack
		userRepo = persistence.NewSQLCUserRepository(db, logger)
		userService = services.NewUserService(userRepo)
		_ = handlers.NewUserHandler(userService, logger)
	})

	AfterEach(func() {
		cleanupDB(db)
	})

	Describe("🏗️ Full Stack Integration", func() {
		Context("complete user lifecycle through all layers", func() {
			It("should handle user creation through domain -> infrastructure -> persistence", func() {
				// Test the complete flow: Service -> Repository -> Database
				id := createTestUserID("integration-user-1")
				email := "integration@example.com"
				name := "Integration Test User"

				// Create user through service (which uses repository -> database)
				createdUser, err := userService.CreateUser(ctx, id, email, name)

				Expect(err).ToNot(HaveOccurred())
				Expect(createdUser).ToNot(BeNil())
				Expect(createdUser.ID).To(Equal(id))
				Expect(createdUser.Email).To(Equal(email))
				Expect(createdUser.Name).To(Equal(name))

				// Verify data persistence at database level
				var dbEmail, dbName string
				var dbCreated, dbModified time.Time
				query := "SELECT email, name, created, modified FROM users WHERE id = ?"
				err = db.QueryRowContext(ctx, query, id.String()).Scan(&dbEmail, &dbName, &dbCreated, &dbModified)

				Expect(err).ToNot(HaveOccurred())
				Expect(dbEmail).To(Equal(email))
				Expect(dbName).To(Equal(name))
				Expect(dbCreated).ToNot(BeZero())
				Expect(dbModified).ToNot(BeZero())
			})

			It("should handle user retrieval through persistence -> infrastructure -> domain", func() {
				// First create a user directly in database
				id := createTestUserID("integration-user-2")
				email := "retrieve@example.com"
				name := "Retrieve Test User"
				now := time.Now()

				_, err := db.ExecContext(ctx,
					"INSERT INTO users (id, email, name, created, modified) VALUES (?, ?, ?, ?, ?)",
					id.String(), email, name, now, now)
				Expect(err).ToNot(HaveOccurred())

				// Retrieve through service layer
				user, err := userService.GetUser(ctx, id)

				Expect(err).ToNot(HaveOccurred())
				Expect(user).ToNot(BeNil())
				Expect(user.ID).To(Equal(id))
				Expect(user.Email).To(Equal(email))
				Expect(user.Name).To(Equal(name))
				Expect(user.Created).To(BeTemporally("~", now, time.Second))
				Expect(user.Modified).To(BeTemporally("~", now, time.Second))
			})

			It("should handle user updates across all layers", func() {
				// Create user
				id := createTestUserID("integration-user-3")
				originalEmail := "original@example.com"
				originalName := "Original Name"

				user, err := userService.CreateUser(ctx, id, originalEmail, originalName)
				Expect(err).ToNot(HaveOccurred())

				// Update user
				newEmail := "updated@example.com"
				newName := "Updated Name"

				updatedUser, err := userService.UpdateUser(ctx, id, newEmail, newName)
				Expect(err).ToNot(HaveOccurred())

				// Verify update propagated to database
				var dbEmail, dbName string
				var dbModified time.Time
				query := "SELECT email, name, modified FROM users WHERE id = ?"
				err = db.QueryRowContext(ctx, query, id.String()).Scan(&dbEmail, &dbName, &dbModified)

				Expect(err).ToNot(HaveOccurred())
				Expect(dbEmail).To(Equal(newEmail))
				Expect(dbName).To(Equal(newName))
				Expect(dbModified).To(BeTemporally(">", user.Created))

				// Verify consistency across layers
				Expect(updatedUser.Email).To(Equal(dbEmail))
				Expect(updatedUser.Name).To(Equal(dbName))
			})

			It("should handle user deletion across all layers", func() {
				// Create user
				id := createTestUserID("integration-user-4")
				_, err := userService.CreateUser(ctx, id, "delete@example.com", "Delete User")
				Expect(err).ToNot(HaveOccurred())

				// Verify user exists in database
				var count int
				err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = ?", id.String()).Scan(&count)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(1))

				// Delete user
				err = userService.DeleteUser(ctx, id)
				Expect(err).ToNot(HaveOccurred())

				// Verify deletion in database
				err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = ?", id.String()).Scan(&count)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(0))

				// Verify deletion through service layer
				user, err := userService.GetUser(ctx, id)
				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("🔄 Cross-Layer Data Consistency", func() {
		Context("with multiple operations", func() {
			It("should maintain consistency across complex operation sequences", func() {
				const numUsers = 5
				userIDs := make([]values.UserID, numUsers)

				// Create multiple users
				for i := 0; i < numUsers; i++ {
					id := createTestUserID(fmt.Sprintf("consistency-user-%d", i))
					email := fmt.Sprintf("consistency%d@example.com", i)
					name := fmt.Sprintf("Consistency User %d", i)

					user, err := userService.CreateUser(ctx, id, email, name)
					Expect(err).ToNot(HaveOccurred())
					userIDs[i] = id

					// Verify immediate consistency
					retrievedUser, err := userService.GetUser(ctx, id)
					Expect(err).ToNot(HaveOccurred())
					Expect(retrievedUser.ID).To(Equal(user.ID))
					Expect(retrievedUser.Email).To(Equal(user.Email))
					Expect(retrievedUser.Name).To(Equal(user.Name))
				}

				// Verify list operation consistency
				users, err := userService.ListUsers(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(numUsers))

				// Verify database count consistency
				var dbCount int
				err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&dbCount)
				Expect(err).ToNot(HaveOccurred())
				Expect(dbCount).To(Equal(numUsers))

				// Update some users
				for i := 0; i < numUsers/2; i++ {
					newEmail := fmt.Sprintf("updated-consistency%d@example.com", i)
					newName := fmt.Sprintf("Updated Consistency User %d", i)

					updatedUser, err := userService.UpdateUser(ctx, userIDs[i], newEmail, newName)
					Expect(err).ToNot(HaveOccurred())

					// Verify update consistency
					retrievedUser, err := userService.GetUser(ctx, userIDs[i])
					Expect(err).ToNot(HaveOccurred())
					Expect(retrievedUser.Email).To(Equal(updatedUser.Email))
					Expect(retrievedUser.Name).To(Equal(updatedUser.Name))
					Expect(retrievedUser.Modified).To(BeTemporally("~", updatedUser.Modified, time.Millisecond))
				}

				// Delete some users
				for i := numUsers / 2; i < numUsers; i++ {
					err := userService.DeleteUser(ctx, userIDs[i])
					Expect(err).ToNot(HaveOccurred())

					// Verify deletion consistency
					user, err := userService.GetUser(ctx, userIDs[i])
					Expect(user).To(BeNil())
					Expect(err).To(HaveOccurred())
				}

				// Verify final state consistency
				finalUsers, err := userService.ListUsers(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(finalUsers).To(HaveLen(numUsers / 2))

				err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&dbCount)
				Expect(err).ToNot(HaveOccurred())
				Expect(dbCount).To(Equal(len(finalUsers)))
			})
		})
	})

	Describe("🎭 Domain Logic Integration", func() {
		Context("business rules enforcement across layers", func() {
			It("should enforce email uniqueness constraint across all layers", func() {
				email := "unique@example.com"

				// Create first user
				id1 := createTestUserID("unique-user-1")
				user1, err := userService.CreateUser(ctx, id1, email, "First User")
				Expect(err).ToNot(HaveOccurred())
				Expect(user1).ToNot(BeNil())

				// Attempt to create second user with same email
				id2 := createTestUserID("unique-user-2")
				user2, err := userService.CreateUser(ctx, id2, email, "Second User")

				Expect(user2).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserAlreadyExists))

				// Verify only one user exists in database
				var count int
				err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(1))

				// Verify through service layer
				users, err := userService.ListUsers(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(1))
				Expect(users[0].Email).To(Equal(email))
				Expect(users[0].Name).To(Equal("First User"))
			})

			It("should enforce domain validation rules across layers", func() {
				invalidTestCases := []struct {
					email string
					name  string
					desc  string
				}{
					{"invalid-email", "Valid Name", "invalid email format"},
					{"valid@example.com", "", "empty name"},
					{"valid@example.com", "A", "too short name"},
					{"valid@example.com", "123", "numeric only name"},
				}

				for _, tc := range invalidTestCases {
					id := createTestUserID("validation-test")
					user, err := userService.CreateUser(ctx, id, tc.email, tc.name)

					Expect(user).To(BeNil(), "should reject: %s", tc.desc)
					Expect(err).To(HaveOccurred(), "should return error for: %s", tc.desc)

					// Verify no data was persisted
					var count int
					err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = ?", id.String()).Scan(&count)
					Expect(err).ToNot(HaveOccurred())
					Expect(count).To(Equal(0), "should not persist invalid data: %s", tc.desc)
				}
			})
		})
	})

	Describe("🚨 Error Propagation Integration", func() {
		Context("error handling across layer boundaries", func() {
			It("should properly propagate repository errors through service to application layer", func() {
				// Close database to simulate connection error
				err := db.Close()
				Expect(err).ToNot(HaveOccurred())

				// Attempt operation that will fail
				id := createTestUserID("error-test-user")
				user, err := userService.CreateUser(ctx, id, "error@example.com", "Error User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("sql: database is closed"))
			})

			It("should handle context cancellation across all layers", func() {
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				id := createTestUserID("cancelled-user")
				user, err := userService.CreateUser(cancelledCtx, id, "cancelled@example.com", "Cancelled User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("should handle timeouts gracefully across layers", func() {
				timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()

				// Add delay to ensure timeout
				time.Sleep(10 * time.Nanosecond)

				id := createTestUserID("timeout-user")
				user, err := userService.CreateUser(timeoutCtx, id, "timeout@example.com", "Timeout User")

				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("🔀 Transaction Boundaries", func() {
		Context("with multiple repository operations", func() {
			It("should maintain atomicity for complex operations", func() {
				// This tests that individual operations are atomic
				// Even though we don't have explicit transactions in this simple example,
				// each repository operation should be atomic

				id := createTestUserID("atomic-user")
				email := "atomic@example.com"
				name := "Atomic User"

				// Create user
				_, err := userService.CreateUser(ctx, id, email, name)
				Expect(err).ToNot(HaveOccurred())

				// Update user
				newEmail := "updated-atomic@example.com"
				newName := "Updated Atomic User"

				updatedUser, err := userService.UpdateUser(ctx, id, newEmail, newName)
				Expect(err).ToNot(HaveOccurred())

				// Verify consistency - no partial updates
				retrievedUser, err := userService.GetUser(ctx, id)
				Expect(err).ToNot(HaveOccurred())

				// Both email and name should be updated, not just one
				Expect(retrievedUser.Email).To(Equal(newEmail))
				Expect(retrievedUser.Name).To(Equal(newName))
				Expect(retrievedUser.Email).To(Equal(updatedUser.Email))
				Expect(retrievedUser.Name).To(Equal(updatedUser.Name))

				// Verify in database
				var dbEmail, dbName string
				err = db.QueryRowContext(ctx, "SELECT email, name FROM users WHERE id = ?", id.String()).Scan(&dbEmail, &dbName)
				Expect(err).ToNot(HaveOccurred())
				Expect(dbEmail).To(Equal(newEmail))
				Expect(dbName).To(Equal(newName))
			})
		})
	})

	Describe("🔍 Performance Integration", func() {
		Context("with realistic data volumes", func() {
			It("should perform efficiently with moderate data load", func() {
				const numUsers = 100

				startTime := time.Now()

				// Create users
				for i := 0; i < numUsers; i++ {
					id := createTestUserID(fmt.Sprintf("perf-user-%d", i))
					email := fmt.Sprintf("perf%d@example.com", i)
					name := fmt.Sprintf("Performance User %d", i)

					_, err := userService.CreateUser(ctx, id, email, name)
					Expect(err).ToNot(HaveOccurred())
				}

				createDuration := time.Since(startTime)

				// List all users
				listStart := time.Now()
				users, err := userService.ListUsers(ctx)
				listDuration := time.Since(listStart)

				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(numUsers))

				// Performance expectations (basic smoke test)
				createOpsPerSec := float64(numUsers) / createDuration.Seconds()
				Expect(createOpsPerSec).To(BeNumerically(">", 50), "create operations should be reasonably fast")

				Expect(listDuration).To(BeNumerically("<", time.Second), "list operation should complete quickly")

				By(fmt.Sprintf("Create performance: %.2f ops/sec, List duration: %v",
					createOpsPerSec, listDuration))
			})
		})
	})
})
