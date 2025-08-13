package persistence_test

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

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/internal/infrastructure/persistence"
)

func TestUserRepositorySQL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserRepositorySQL Suite")
}

var _ = Describe("SQLUserRepository", func() {
	var (
		repo   *persistence.SQLUserRepository
		db     *sql.DB
		ctx    context.Context
		logger *slog.Logger
	)

	// Test helper functions
	createTestUserID := func(id string) values.UserID {
		userID, err := values.NewUserID(id)
		Expect(err).ToNot(HaveOccurred())
		return userID
	}

	createTestUser := func(idSuffix, email, name string) *entities.User {
		id := createTestUserID(idSuffix)
		user, err := entities.NewUser(id, email, name)
		Expect(err).ToNot(HaveOccurred())
		return user
	}

	createAndSaveTestUser := func(idSuffix, email, name string) *entities.User {
		user := createTestUser(idSuffix, email, name)
		err := repo.Save(ctx, user)
		Expect(err).ToNot(HaveOccurred())
		return user
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
		repo = persistence.NewSQLUserRepository(db, logger)
	})

	AfterEach(func() {
		cleanupDB(db)
	})

	Describe("NewSQLUserRepository", func() {
		Context("with valid parameters", func() {
			It("should create a new repository successfully", func() {
				testDB := setupInMemoryDB()
				defer cleanupDB(testDB)

				testRepo := persistence.NewSQLUserRepository(testDB, logger)

				Expect(testRepo).ToNot(BeNil())

				// Verify schema was initialized by attempting to query the users table
				var count int
				err := testDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(0))
			})
		})

		Context("with nil database", func() {
			It("should create repository but operations should fail", func() {
				// This tests that we don't panic with nil database during construction
				testRepo := persistence.NewSQLUserRepository(nil, logger)
				Expect(testRepo).ToNot(BeNil())

				// But operations should fail with meaningful errors
				user := createTestUser("test-user", "test@example.com", "Test User")
				err := testRepo.Save(ctx, user)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("database connection is nil"))
			})
		})

		Context("with nil logger", func() {
			It("should handle nil logger gracefully", func() {
				testDB := setupInMemoryDB()
				defer cleanupDB(testDB)

				testRepo := persistence.NewSQLUserRepository(testDB, nil)
				Expect(testRepo).ToNot(BeNil())

				// Should still work with nil logger
				user := createTestUser("test-user", "test@example.com", "Test User")
				err := testRepo.Save(ctx, user)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Save", func() {
		Context("with valid user", func() {
			It("should save user successfully", func() {
				user := createTestUser("test-user-1", "test@example.com", "Test User")

				err := repo.Save(ctx, user)

				Expect(err).ToNot(HaveOccurred())

				// Verify user was saved by retrieving it
				savedUser, err := repo.FindByID(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(savedUser.ID).To(Equal(user.ID))
				Expect(savedUser.Email).To(Equal(user.Email))
				Expect(savedUser.Name).To(Equal(user.Name))
			})

			It("should update existing user on duplicate save", func() {
				user := createAndSaveTestUser("test-user-1", "original@example.com", "Original User")

				// Update user details
				err := user.SetEmail("updated@example.com")
				Expect(err).ToNot(HaveOccurred())
				err = user.SetName("Updated User")
				Expect(err).ToNot(HaveOccurred())

				// Save updated user
				err = repo.Save(ctx, user)
				Expect(err).ToNot(HaveOccurred())

				// Verify update
				savedUser, err := repo.FindByID(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(savedUser.Email).To(Equal("updated@example.com"))
				Expect(savedUser.Name).To(Equal("Updated User"))
			})
		})

		Context("with nil user", func() {
			It("should return error for nil user", func() {
				err := repo.Save(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("user cannot be nil"))
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				user := createTestUser("test-user-1", "test@example.com", "Test User")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				err := repo.Save(cancelledCtx, user)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("with database constraints", func() {
			It("should handle unique email constraint", func() {
				user1 := createAndSaveTestUser("user-1", "same@example.com", "User One")
				user2 := createTestUser("user-2", "same@example.com", "User Two")

				err := repo.Save(ctx, user2)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to save user"))

				// Verify first user is still there
				savedUser, err := repo.FindByID(ctx, user1.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(savedUser.Name).To(Equal("User One"))
			})
		})
	})

	Describe("FindByID", func() {
		Context("when user exists", func() {
			It("should return the user", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")

				foundUser, err := repo.FindByID(ctx, user.ID)

				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser).ToNot(BeNil())
				Expect(foundUser.ID).To(Equal(user.ID))
				Expect(foundUser.Email).To(Equal(user.Email))
				Expect(foundUser.Name).To(Equal(user.Name))
				Expect(foundUser.Created).To(BeTemporally("~", user.Created, time.Second))
				Expect(foundUser.Modified).To(BeTemporally("~", user.Modified, time.Second))
			})

			It("should return user with all fields populated", func() {
				user := createAndSaveTestUser("test-user-full", "full@example.com", "Full Test User")

				foundUser, err := repo.FindByID(ctx, user.ID)

				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser.ID.IsEmpty()).To(BeFalse())
				Expect(foundUser.Email).ToNot(BeEmpty())
				Expect(foundUser.Name).ToNot(BeEmpty())
				Expect(foundUser.Created.IsZero()).To(BeFalse())
				Expect(foundUser.Modified.IsZero()).To(BeFalse())
			})
		})

		Context("when user does not exist", func() {
			It("should return ErrUserNotFound", func() {
				nonExistentID := createTestUserID("nonexistent-user")

				user, err := repo.FindByID(ctx, nonExistentID)

				Expect(user).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})

			It("should return ErrUserNotFound for empty ID", func() {
				emptyID := values.UserID{}

				user, err := repo.FindByID(ctx, emptyID)

				Expect(user).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				savedUser := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				foundUser, err := repo.FindByID(cancelledCtx, savedUser.ID)

				Expect(foundUser).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with timeout context", func() {
			It("should handle timeout gracefully", func() {
				savedUser := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()

				// Add small delay to ensure timeout
				time.Sleep(10 * time.Nanosecond)

				foundUser, err := repo.FindByID(timeoutCtx, savedUser.ID)

				Expect(foundUser).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("FindByEmail", func() {
		Context("when user exists", func() {
			It("should return the user", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")

				foundUser, err := repo.FindByEmail(ctx, user.Email)

				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser).ToNot(BeNil())
				Expect(foundUser.ID).To(Equal(user.ID))
				Expect(foundUser.Email).To(Equal(user.Email))
				Expect(foundUser.Name).To(Equal(user.Name))
			})

			It("should be case sensitive", func() {
				createAndSaveTestUser("test-user-1", "test@example.com", "Test User")

				foundUser, err := repo.FindByEmail(ctx, "TEST@EXAMPLE.COM")

				Expect(foundUser).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})

			It("should handle special email characters", func() {
				createAndSaveTestUser("test-user-special", "user+tag@sub.example.com", "Special User")

				foundUser, err := repo.FindByEmail(ctx, "user+tag@sub.example.com")

				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser).ToNot(BeNil())
				Expect(foundUser.Email).To(Equal("user+tag@sub.example.com"))
			})
		})

		Context("when user does not exist", func() {
			It("should return ErrUserNotFound", func() {
				user, err := repo.FindByEmail(ctx, "nonexistent@example.com")

				Expect(user).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})

			It("should return ErrUserNotFound for empty email", func() {
				user, err := repo.FindByEmail(ctx, "")

				Expect(user).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})

			It("should return ErrUserNotFound for malformed email", func() {
				user, err := repo.FindByEmail(ctx, "invalid-email")

				Expect(user).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})
		})

		Context("with multiple users", func() {
			It("should return the correct user", func() {
				user1 := createAndSaveTestUser("user-1", "user1@example.com", "User One")
				user2 := createAndSaveTestUser("user-2", "user2@example.com", "User Two")

				foundUser, err := repo.FindByEmail(ctx, "user2@example.com")

				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser.ID).To(Equal(user2.ID))
				Expect(foundUser.Email).To(Equal("user2@example.com"))
				Expect(foundUser.Name).To(Equal("User Two"))
				Expect(foundUser.ID).ToNot(Equal(user1.ID))
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				foundUser, err := repo.FindByEmail(cancelledCtx, "test@example.com")

				Expect(foundUser).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Delete", func() {
		Context("when user exists", func() {
			It("should delete user successfully", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")

				err := repo.Delete(ctx, user.ID)

				Expect(err).ToNot(HaveOccurred())

				// Verify user is deleted
				foundUser, err := repo.FindByID(ctx, user.ID)
				Expect(foundUser).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})

			It("should delete only the specified user", func() {
				user1 := createAndSaveTestUser("user-1", "user1@example.com", "User One")
				user2 := createAndSaveTestUser("user-2", "user2@example.com", "User Two")

				err := repo.Delete(ctx, user1.ID)

				Expect(err).ToNot(HaveOccurred())

				// Verify user1 is deleted
				foundUser1, err := repo.FindByID(ctx, user1.ID)
				Expect(foundUser1).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))

				// Verify user2 still exists
				foundUser2, err := repo.FindByID(ctx, user2.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser2.ID).To(Equal(user2.ID))
			})

			It("should handle concurrent deletes gracefully", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")

				// First delete should succeed
				err1 := repo.Delete(ctx, user.ID)
				Expect(err1).ToNot(HaveOccurred())

				// Second delete should return ErrUserNotFound
				err2 := repo.Delete(ctx, user.ID)
				Expect(err2).To(Equal(repositories.ErrUserNotFound))
			})
		})

		Context("when user does not exist", func() {
			It("should return ErrUserNotFound", func() {
				nonExistentID := createTestUserID("nonexistent-user")

				err := repo.Delete(ctx, nonExistentID)

				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})

			It("should return ErrUserNotFound for empty ID", func() {
				emptyID := values.UserID{}

				err := repo.Delete(ctx, emptyID)

				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				err := repo.Delete(cancelledCtx, user.ID)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("List", func() {
		Context("when users exist", func() {
			It("should return all users", func() {
				user1 := createAndSaveTestUser("user-1", "user1@example.com", "User One")
				user2 := createAndSaveTestUser("user-2", "user2@example.com", "User Two")

				users, err := repo.List(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(2))

				// Check that both users are present
				userIDs := make([]values.UserID, len(users))
				for i, u := range users {
					userIDs[i] = u.ID
				}
				Expect(userIDs).To(ContainElement(user1.ID))
				Expect(userIDs).To(ContainElement(user2.ID))
			})

			It("should return users ordered by creation time", func() {
				// Create users with slight delay to ensure different creation times
				user1 := createAndSaveTestUser("user-1", "user1@example.com", "User One")
				time.Sleep(10 * time.Millisecond)
				user2 := createAndSaveTestUser("user-2", "user2@example.com", "User Two")

				users, err := repo.List(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(2))

				// Should be ordered by creation time (ASC)
				Expect(users[0].ID).To(Equal(user1.ID))
				Expect(users[1].ID).To(Equal(user2.ID))
			})

			It("should return complete user data", func() {
				createAndSaveTestUser("user-1", "user1@example.com", "User One")

				users, err := repo.List(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(1))

				user := users[0]
				Expect(user.ID.IsEmpty()).To(BeFalse())
				Expect(user.Email).ToNot(BeEmpty())
				Expect(user.Name).ToNot(BeEmpty())
				Expect(user.Created.IsZero()).To(BeFalse())
				Expect(user.Modified.IsZero()).To(BeFalse())
			})

			It("should handle large number of users", func() {
				const numUsers = 100

				// Create many users
				for i := 0; i < numUsers; i++ {
					createAndSaveTestUser(
						fmt.Sprintf("user-%d", i),
						fmt.Sprintf("user%d@example.com", i),
						fmt.Sprintf("User %d", i),
					)
				}

				users, err := repo.List(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(numUsers))
			})
		})

		Context("when no users exist", func() {
			It("should return empty slice", func() {
				users, err := repo.List(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(0))
				Expect(users).ToNot(BeNil())
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				users, err := repo.List(cancelledCtx)

				Expect(users).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with database errors", func() {
			It("should handle database connection loss gracefully", func() {
				createAndSaveTestUser("test-user-1", "test@example.com", "Test User")

				// Close database to simulate connection loss
				err := db.Close()
				Expect(err).ToNot(HaveOccurred())

				users, err := repo.List(ctx)

				Expect(users).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("FindByUsername", func() {
		Context("when user exists", func() {
			It("should return the user", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "TestUsername")

				foundUser, err := repo.FindByUsername(ctx, user.Name)

				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser).ToNot(BeNil())
				Expect(foundUser.ID).To(Equal(user.ID))
				Expect(foundUser.Email).To(Equal(user.Email))
				Expect(foundUser.Name).To(Equal(user.Name))
			})

			It("should be case sensitive", func() {
				createAndSaveTestUser("test-user-1", "test@example.com", "TestUser")

				foundUser, err := repo.FindByUsername(ctx, "testuser")

				Expect(foundUser).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})

			It("should handle special characters in username", func() {
				createAndSaveTestUser("test-user-special", "special@example.com", "Test-User_123")

				foundUser, err := repo.FindByUsername(ctx, "Test-User_123")

				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser).ToNot(BeNil())
				Expect(foundUser.Name).To(Equal("Test-User_123"))
			})
		})

		Context("when user does not exist", func() {
			It("should return ErrUserNotFound", func() {
				user, err := repo.FindByUsername(ctx, "nonexistent-user")

				Expect(user).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})

			It("should return ErrUserNotFound for empty username", func() {
				user, err := repo.FindByUsername(ctx, "")

				Expect(user).To(BeNil())
				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})
		})

		Context("with multiple users with similar names", func() {
			It("should return the exact match", func() {
				user1 := createAndSaveTestUser("user-1", "user1@example.com", "TestUser")
				user2 := createAndSaveTestUser("user-2", "user2@example.com", "TestUserExtended")

				foundUser, err := repo.FindByUsername(ctx, "TestUser")

				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser.ID).To(Equal(user1.ID))
				Expect(foundUser.Name).To(Equal("TestUser"))
				Expect(foundUser.ID).ToNot(Equal(user2.ID))
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				createAndSaveTestUser("test-user-1", "test@example.com", "TestUser")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				foundUser, err := repo.FindByUsername(cancelledCtx, "TestUser")

				Expect(foundUser).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Update", func() {
		Context("when user exists", func() {
			It("should update user successfully", func() {
				user := createAndSaveTestUser("test-user-1", "original@example.com", "Original Name")

				// Update user details
				err := user.SetEmail("updated@example.com")
				Expect(err).ToNot(HaveOccurred())
				err = user.SetName("Updated Name")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Update(ctx, user)

				Expect(err).ToNot(HaveOccurred())

				// Verify update
				updatedUser, err := repo.FindByID(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedUser.Email).To(Equal("updated@example.com"))
				Expect(updatedUser.Name).To(Equal("Updated Name"))
				Expect(updatedUser.Modified).To(BeTemporally(">", updatedUser.Created))
			})

			It("should preserve unchanged fields", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				originalCreated := user.Created

				// Update only email
				err := user.SetEmail("newemail@example.com")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Update(ctx, user)
				Expect(err).ToNot(HaveOccurred())

				// Verify update
				updatedUser, err := repo.FindByID(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(updatedUser.Email).To(Equal("newemail@example.com"))
				Expect(updatedUser.Name).To(Equal("Test User"))                                 // Unchanged
				Expect(updatedUser.Created).To(BeTemporally("~", originalCreated, time.Second)) // Unchanged
			})

			It("should handle concurrent updates", func() {
				user := createAndSaveTestUser("concurrent-user", "concurrent@example.com", "Concurrent User")

				// Create two copies for concurrent updates
				user1 := *user
				user2 := *user

				err := user1.SetEmail("update1@example.com")
				Expect(err).ToNot(HaveOccurred())
				err = user2.SetEmail("update2@example.com")
				Expect(err).ToNot(HaveOccurred())

				// First update should succeed
				err = repo.Update(ctx, &user1)
				Expect(err).ToNot(HaveOccurred())

				// Second update should also succeed (last write wins)
				err = repo.Update(ctx, &user2)
				Expect(err).ToNot(HaveOccurred())

				// Verify final state
				finalUser, err := repo.FindByID(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(finalUser.Email).To(Equal("update2@example.com"))
			})
		})

		Context("when user does not exist", func() {
			It("should return ErrUserNotFound", func() {
				nonExistentUser := createTestUser("nonexistent-user", "test@example.com", "Test User")

				err := repo.Update(ctx, nonExistentUser)

				Expect(err).To(Equal(repositories.ErrUserNotFound))
			})
		})

		Context("with nil user", func() {
			It("should return error for nil user", func() {
				err := repo.Update(ctx, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("user cannot be nil"))
			})
		})

		Context("with database constraints", func() {
			It("should handle unique email constraint violation", func() {
				user1 := createAndSaveTestUser("user-1", "user1@example.com", "User One")
				user2 := createAndSaveTestUser("user-2", "user2@example.com", "User Two")

				// Try to update user2 with user1's email
				err := user2.SetEmail("user1@example.com")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Update(ctx, user2)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to update user"))

				// Verify user1 is unchanged
				foundUser1, err := repo.FindByID(ctx, user1.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser1.Email).To(Equal("user1@example.com"))
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				err := user.SetEmail("updated@example.com")
				Expect(err).ToNot(HaveOccurred())

				err = repo.Update(cancelledCtx, user)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Count", func() {
		Context("when users exist", func() {
			It("should return correct count", func() {
				createAndSaveTestUser("user-1", "user1@example.com", "User One")
				createAndSaveTestUser("user-2", "user2@example.com", "User Two")
				createAndSaveTestUser("user-3", "user3@example.com", "User Three")

				count, err := repo.Count(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(3))
			})

			It("should update count after operations", func() {
				// Initial count
				count, err := repo.Count(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(0))

				// Add users
				user1 := createAndSaveTestUser("user-1", "user1@example.com", "User One")
				createAndSaveTestUser("user-2", "user2@example.com", "User Two")

				count, err = repo.Count(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(2))

				// Delete user
				err = repo.Delete(ctx, user1.ID)
				Expect(err).ToNot(HaveOccurred())

				count, err = repo.Count(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(1))
			})

			It("should handle large counts", func() {
				const numUsers = 50

				// Create many users
				for i := 0; i < numUsers; i++ {
					createAndSaveTestUser(
						fmt.Sprintf("user-%d", i),
						fmt.Sprintf("user%d@example.com", i),
						fmt.Sprintf("User %d", i),
					)
				}

				count, err := repo.Count(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(numUsers))
			})
		})

		Context("when no users exist", func() {
			It("should return zero", func() {
				count, err := repo.Count(ctx)

				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(0))
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				count, err := repo.Count(cancelledCtx)

				Expect(count).To(Equal(0))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Exists", func() {
		Context("when user exists", func() {
			It("should return true", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")

				exists, err := repo.Exists(ctx, user.ID)

				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(BeTrue())
			})

			It("should return true for multiple users", func() {
				user1 := createAndSaveTestUser("user-1", "user1@example.com", "User One")
				user2 := createAndSaveTestUser("user-2", "user2@example.com", "User Two")

				exists1, err := repo.Exists(ctx, user1.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists1).To(BeTrue())

				exists2, err := repo.Exists(ctx, user2.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists2).To(BeTrue())
			})
		})

		Context("when user does not exist", func() {
			It("should return false", func() {
				nonExistentID := createTestUserID("nonexistent-user")

				exists, err := repo.Exists(ctx, nonExistentID)

				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(BeFalse())
			})

			It("should return false for empty ID", func() {
				emptyID := values.UserID{}

				exists, err := repo.Exists(ctx, emptyID)

				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})

		Context("after user operations", func() {
			It("should return false after deletion", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")

				// Verify exists before deletion
				exists, err := repo.Exists(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(BeTrue())

				// Delete user
				err = repo.Delete(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())

				// Verify doesn't exist after deletion
				exists, err = repo.Exists(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(BeFalse())
			})

			It("should return true after creation", func() {
				user := createTestUser("new-user", "new@example.com", "New User")

				// Verify doesn't exist before creation
				exists, err := repo.Exists(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(BeFalse())

				// Save user
				err = repo.Save(ctx, user)
				Expect(err).ToNot(HaveOccurred())

				// Verify exists after creation
				exists, err = repo.Exists(ctx, user.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(BeTrue())
			})
		})

		Context("with cancelled context", func() {
			It("should return context error", func() {
				user := createAndSaveTestUser("test-user-1", "test@example.com", "Test User")
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel()

				exists, err := repo.Exists(cancelledCtx, user.ID)

				Expect(exists).To(BeFalse())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	// Additional edge case tests
	Describe("Edge Cases and Error Handling", func() {
		Context("database schema validation", func() {
			It("should handle missing tables gracefully", func() {
				// Create a new database without schema initialization
				testDB := setupInMemoryDB()
				defer cleanupDB(testDB)

				// Drop the users table to simulate missing schema
				_, err := testDB.ExecContext(ctx, "DROP TABLE IF EXISTS users")
				Expect(err).ToNot(HaveOccurred())

				testRepo := persistence.NewSQLUserRepository(testDB, logger)
				user := createTestUser("test-user-1", "test@example.com", "Test User")

				// Should be able to save after schema initialization
				err = testRepo.Save(ctx, user)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("transaction boundaries", func() {
			It("should handle concurrent operations safely", func() {
				user1 := createTestUser("concurrent-user-1", "concurrent1@example.com", "Concurrent User 1")
				user2 := createTestUser("concurrent-user-2", "concurrent2@example.com", "Concurrent User 2")

				// Simulate concurrent saves using the shared repo with initialized schema
				done := make(chan error, 2)

				go func() {
					done <- repo.Save(ctx, user1)
				}()

				go func() {
					done <- repo.Save(ctx, user2)
				}()

				// Wait for both operations
				err1 := <-done
				err2 := <-done

				Expect(err1).ToNot(HaveOccurred())
				Expect(err2).ToNot(HaveOccurred())

				// Verify both users exist
				foundUser1, err := repo.FindByID(ctx, user1.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser1.ID).To(Equal(user1.ID))

				foundUser2, err := repo.FindByID(ctx, user2.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(foundUser2.ID).To(Equal(user2.ID))
			})
		})

		Context("data integrity", func() {
			It("should preserve user data across operations", func() {
				originalUser := createAndSaveTestUser("integrity-user", "integrity@example.com", "Integrity User")

				// Perform various operations
				foundUser, err := repo.FindByID(ctx, originalUser.ID)
				Expect(err).ToNot(HaveOccurred())

				foundByEmail, err := repo.FindByEmail(ctx, originalUser.Email)
				Expect(err).ToNot(HaveOccurred())

				users, err := repo.List(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(1))

				// All operations should return consistent data
				Expect(foundUser.ID).To(Equal(originalUser.ID))
				Expect(foundUser.Email).To(Equal(originalUser.Email))
				Expect(foundUser.Name).To(Equal(originalUser.Name))

				Expect(foundByEmail.ID).To(Equal(originalUser.ID))
				Expect(users[0].ID).To(Equal(originalUser.ID))
			})
		})
	})
})
