package services_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

func TestUserServiceConcurrency(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "🔄 UserService Concurrent Access Testing Suite")
}

var _ = Describe("🔄 UserService Concurrent Access Testing", func() {
	var (
		userService *services.UserService
		userRepo    repositories.UserRepository
		ctx         context.Context
	)

	// Test helper functions
	createTestUserID := func(id string) values.UserID {
		userID, err := values.NewUserID(id)
		Expect(err).ToNot(HaveOccurred())
		return userID
	}

	BeforeEach(func() {
		ctx = context.Background()
		userRepo = repositories.NewInMemoryUserRepository()
		userService = services.NewUserService(userRepo)
	})

	Describe("🏃‍♂️ Concurrent User Creation", func() {
		Context("with different users", func() {
			It("should handle multiple concurrent user creations successfully", func() {
				const numGoroutines = 10
				const numUsersPerGoroutine = 5

				var wg sync.WaitGroup
				results := make(chan error, numGoroutines*numUsersPerGoroutine)
				userCounts := make(chan int, numGoroutines)

				// Create multiple goroutines creating users concurrently
				for g := 0; g < numGoroutines; g++ {
					wg.Add(1)
					go func(goroutineID int) {
						defer wg.Done()
						localSuccessCount := 0

						for u := 0; u < numUsersPerGoroutine; u++ {
							id := createTestUserID(fmt.Sprintf("concurrent-user-g%d-u%d", goroutineID, u))
							email := fmt.Sprintf("user-g%d-u%d@example.com", goroutineID, u)
							name := fmt.Sprintf("User G%d U%d", goroutineID, u)

							_, err := userService.CreateUser(ctx, id, email, name)
							results <- err
							if err == nil {
								localSuccessCount++
							}
						}
						userCounts <- localSuccessCount
					}(g)
				}

				wg.Wait()
				close(results)
				close(userCounts)

				// Verify all operations succeeded
				successCount := 0
				errorCount := 0
				for err := range results {
					if err == nil {
						successCount++
					} else {
						errorCount++
					}
				}

				Expect(successCount).To(Equal(numGoroutines * numUsersPerGoroutine))
				Expect(errorCount).To(Equal(0))

				// Verify total user count
				finalUsers, err := userService.ListUsers(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(finalUsers).To(HaveLen(numGoroutines * numUsersPerGoroutine))
			})
		})

		Context("with duplicate emails", func() {
			It("should handle concurrent creation attempts with same email", func() {
				const numGoroutines = 10
				const sameEmail = "duplicate@example.com"
				const sameName = "Duplicate User"

				var wg sync.WaitGroup
				results := make(chan error, numGoroutines)

				// Multiple goroutines trying to create users with same email
				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						id := createTestUserID(fmt.Sprintf("duplicate-user-%d", index))
						_, err := userService.CreateUser(ctx, id, sameEmail, sameName)
						results <- err
					}(i)
				}

				wg.Wait()
				close(results)

				// Count successes and failures
				successCount := 0
				conflictCount := 0
				otherErrorCount := 0

				for err := range results {
					if err == nil {
						successCount++
					} else if errors.Is(err, repositories.ErrUserAlreadyExists) {
						conflictCount++
					} else {
						otherErrorCount++
					}
				}

				// Should have exactly one success and the rest conflicts
				Expect(successCount).To(Equal(1), "should create exactly one user")
				Expect(conflictCount).To(Equal(numGoroutines-1), "should reject duplicate attempts")
				Expect(otherErrorCount).To(Equal(0), "should not have other errors")

				// Verify only one user exists
				users, err := userService.ListUsers(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(users).To(HaveLen(1))
				Expect(users[0].GetEmail().String()).To(Equal(sameEmail))
			})
		})
	})

	Describe("🔄 Concurrent Read Operations", func() {
		var testUser *entities.User

		BeforeEach(func() {
			// Create a test user first
			id := createTestUserID("read-test-user")
			var err error
			testUser, err = userService.CreateUser(ctx, id, "read@example.com", "Read Test User")
			Expect(err).ToNot(HaveOccurred())
		})

		Context("with concurrent GetUser calls", func() {
			It("should handle multiple concurrent reads safely", func() {
				const numReaders = 20

				var wg sync.WaitGroup
				results := make(chan error, numReaders)
				userResults := make(chan *entities.User, numReaders)

				// Multiple goroutines reading the same user
				for i := 0; i < numReaders; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						user, err := userService.GetUser(ctx, testUser.ID)
						results <- err
						userResults <- user
					}()
				}

				wg.Wait()
				close(results)
				close(userResults)

				// Verify all reads succeeded
				for err := range results {
					Expect(err).ToNot(HaveOccurred())
				}

				// Verify all reads returned the same user data
				for user := range userResults {
					Expect(user).ToNot(BeNil())
					Expect(user.ID).To(Equal(testUser.ID))
					Expect(user.GetEmail().String()).To(Equal(testUser.GetEmail().String()))
					Expect(user.GetUserName().String()).To(Equal(testUser.GetUserName().String()))
				}
			})
		})

		Context("with concurrent ListUsers calls", func() {
			It("should handle multiple concurrent list operations", func() {
				const numListers = 15

				var wg sync.WaitGroup
				results := make(chan error, numListers)
				listResults := make(chan int, numListers) // Store count of users

				for i := 0; i < numListers; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						users, err := userService.ListUsers(ctx)
						results <- err
						if err == nil {
							listResults <- len(users)
						}
					}()
				}

				wg.Wait()
				close(results)
				close(listResults)

				// Verify all operations succeeded
				for err := range results {
					Expect(err).ToNot(HaveOccurred())
				}

				// Verify all operations returned consistent count
				for count := range listResults {
					Expect(count).To(Equal(1)) // Only our test user exists
				}
			})
		})
	})

	Describe("✏️ Concurrent Update Operations", func() {
		var testUser *entities.User

		BeforeEach(func() {
			// Create a test user for updating
			id := createTestUserID("update-test-user")
			var err error
			testUser, err = userService.CreateUser(ctx, id, "update@example.com", "Update Test User")
			Expect(err).ToNot(HaveOccurred())
		})

		Context("with concurrent updates to same user", func() {
			It("should handle concurrent updates with last-write-wins semantics", func() {
				const numUpdaters = 10

				var wg sync.WaitGroup
				results := make(chan error, numUpdaters)
				updateResults := make(chan string, numUpdaters) // Store final email

				// Multiple goroutines updating the same user
				for i := 0; i < numUpdaters; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						newEmail := fmt.Sprintf("updated%d@example.com", index)
						newName := fmt.Sprintf("Updated User %d", index)

						user, err := userService.UpdateUser(ctx, testUser.ID, newEmail, newName)
						results <- err
						if err == nil {
							updateResults <- user.GetEmail().String()
						}
					}(i)
				}

				wg.Wait()
				close(results)
				close(updateResults)

				// Count successful updates
				successCount := 0
				for err := range results {
					if err == nil {
						successCount++
					}
				}

				// At least some updates should succeed (depending on timing)
				Expect(successCount).To(BeNumerically(">", 0))

				// Verify final state is consistent
				finalUser, err := userService.GetUser(ctx, testUser.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(finalUser.GetEmail().String()).To(MatchRegexp(`updated\d+@example\.com`))
				Expect(finalUser.GetUserName().String()).To(MatchRegexp(`Updated User \d+`))
			})
		})

		Context("with concurrent updates to different users", func() {
			It("should handle updates to different users independently", func() {
				const numUsers = 10

				// Create multiple users
				userIDs := make([]values.UserID, numUsers)
				for i := 0; i < numUsers; i++ {
					id := createTestUserID(fmt.Sprintf("multi-update-user-%d", i))
					email := fmt.Sprintf("multi%d@example.com", i)
					name := fmt.Sprintf("Multi User %d", i)

					_, err := userService.CreateUser(ctx, id, email, name)
					Expect(err).ToNot(HaveOccurred())
					userIDs[i] = id
				}

				var wg sync.WaitGroup
				results := make(chan error, numUsers)

				// Update each user concurrently
				for i := 0; i < numUsers; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						newEmail := fmt.Sprintf("updated-multi%d@example.com", index)
						newName := fmt.Sprintf("Updated Multi User %d", index)

						_, err := userService.UpdateUser(ctx, userIDs[index], newEmail, newName)
						results <- err
					}(i)
				}

				wg.Wait()
				close(results)

				// All updates should succeed
				for err := range results {
					Expect(err).ToNot(HaveOccurred())
				}

				// Verify all users were updated correctly
				for i := 0; i < numUsers; i++ {
					user, err := userService.GetUser(ctx, userIDs[i])
					Expect(err).ToNot(HaveOccurred())
					Expect(user.GetEmail().String()).To(Equal(fmt.Sprintf("updated-multi%d@example.com", i)))
					Expect(user.GetUserName().String()).To(Equal(fmt.Sprintf("Updated Multi User %d", i)))
				}
			})
		})
	})

	Describe("🗑️ Concurrent Delete Operations", func() {
		Context("with concurrent deletes of same user", func() {
			It("should handle multiple delete attempts gracefully", func() {
				// Create a user to delete
				id := createTestUserID("delete-test-user")
				_, err := userService.CreateUser(ctx, id, "delete@example.com", "Delete Test User")
				Expect(err).ToNot(HaveOccurred())

				const numDeleters = 5
				var wg sync.WaitGroup
				results := make(chan error, numDeleters)

				// Multiple goroutines trying to delete the same user
				for i := 0; i < numDeleters; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := userService.DeleteUser(ctx, id)
						results <- err
					}()
				}

				wg.Wait()
				close(results)

				// Count successes and "not found" errors
				successCount := 0
				notFoundCount := 0
				otherErrorCount := 0

				for err := range results {
					if err == nil {
						successCount++
					} else if errors.Is(err, repositories.ErrUserNotFound) {
						notFoundCount++
					} else {
						otherErrorCount++
					}
				}

				// Should have exactly one success and the rest "not found"
				Expect(successCount).To(Equal(1), "should delete exactly once")
				Expect(notFoundCount).To(Equal(numDeleters-1), "should get not found for subsequent attempts")
				Expect(otherErrorCount).To(Equal(0), "should not have other errors")

				// Verify user no longer exists
				user, err := userService.GetUser(ctx, id)
				Expect(user).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with concurrent deletes of different users", func() {
			It("should handle independent deletions correctly", func() {
				const numUsers = 8

				// Create multiple users
				userIDs := make([]values.UserID, numUsers)
				for i := 0; i < numUsers; i++ {
					id := createTestUserID(fmt.Sprintf("multi-delete-user-%d", i))
					email := fmt.Sprintf("multidelete%d@example.com", i)
					name := fmt.Sprintf("Multi Delete User %d", i)

					_, err := userService.CreateUser(ctx, id, email, name)
					Expect(err).ToNot(HaveOccurred())
					userIDs[i] = id
				}

				var wg sync.WaitGroup
				results := make(chan error, numUsers)

				// Delete each user concurrently
				for i := 0; i < numUsers; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						err := userService.DeleteUser(ctx, userIDs[index])
						results <- err
					}(i)
				}

				wg.Wait()
				close(results)

				// All deletions should succeed
				for err := range results {
					Expect(err).ToNot(HaveOccurred())
				}

				// Verify all users are deleted
				finalUsers, err := userService.ListUsers(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(finalUsers).To(HaveLen(0))
			})
		})
	})

	Describe("🔀 Mixed Concurrent Operations", func() {
		Context("with CRUD operations happening simultaneously", func() {
			It("should maintain data consistency across mixed operations", func() {
				const numOperations = 20
				const operationTypes = 4 // Create, Read, Update, Delete

				// Pre-create some users for update/delete operations
				existingUserIDs := make([]values.UserID, 5)
				for i := 0; i < 5; i++ {
					id := createTestUserID(fmt.Sprintf("existing-user-%d", i))
					email := fmt.Sprintf("existing%d@example.com", i)
					name := fmt.Sprintf("Existing User %d", i)

					_, err := userService.CreateUser(ctx, id, email, name)
					Expect(err).ToNot(HaveOccurred())
					existingUserIDs[i] = id
				}

				var wg sync.WaitGroup
				results := make(chan error, numOperations)

				createCount := 0
				readCount := 0
				updateCount := 0
				deleteCount := 0

				// Launch mixed operations
				for i := 0; i < numOperations; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						var err error

						switch index % operationTypes {
						case 0: // Create
							id := createTestUserID(fmt.Sprintf("mixed-create-user-%d", index))
							email := fmt.Sprintf("mixedcreate%d@example.com", index)
							name := fmt.Sprintf("Mixed Create User %d", index)
							_, err = userService.CreateUser(ctx, id, email, name)
							createCount++

						case 1: // Read
							if len(existingUserIDs) > 0 {
								userID := existingUserIDs[index%len(existingUserIDs)]
								_, err = userService.GetUser(ctx, userID)
							}
							readCount++

						case 2: // Update
							if len(existingUserIDs) > 0 {
								userID := existingUserIDs[index%len(existingUserIDs)]
								newEmail := fmt.Sprintf("mixedupdate%d@example.com", index)
								newName := fmt.Sprintf("Mixed Update User %d", index)
								_, err = userService.UpdateUser(ctx, userID, newEmail, newName)
							}
							updateCount++

						case 3: // Delete
							if len(existingUserIDs) > 0 && index < len(existingUserIDs) {
								userID := existingUserIDs[index]
								err = userService.DeleteUser(ctx, userID)
							}
							deleteCount++
						}

						results <- err
					}(i)
				}

				wg.Wait()
				close(results)

				// Count operation results
				successCount := 0
				errorCount := 0
				for err := range results {
					if err == nil {
						successCount++
					} else {
						errorCount++
					}
				}

				// Most operations should succeed (some conflicts expected)
				Expect(successCount).To(BeNumerically(">", numOperations/2))

				// Verify repository is in a consistent state
				finalUsers, err := userService.ListUsers(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(finalUsers).ToNot(BeNil())

				// All remaining users should have valid data
				for _, user := range finalUsers {
					Expect(user.ID.IsEmpty()).To(BeFalse())
					Expect(user.GetEmail().String()).ToNot(BeEmpty())
					Expect(user.GetUserName().String()).ToNot(BeEmpty())
					Expect(user.Created.IsZero()).To(BeFalse())
					Expect(user.Modified.IsZero()).To(BeFalse())
				}
			})
		})
	})

	Describe("⚡ Performance Under Concurrent Load", func() {
		Context("with high concurrency", func() {
			It("should maintain reasonable performance under load", func() {
				const numGoroutines = 50
				const operationsPerGoroutine = 10

				startTime := time.Now()

				var wg sync.WaitGroup
				results := make(chan error, numGoroutines*operationsPerGoroutine)

				// High concurrency load test
				for g := 0; g < numGoroutines; g++ {
					wg.Add(1)
					go func(goroutineID int) {
						defer wg.Done()

						for op := 0; op < operationsPerGoroutine; op++ {
							id := createTestUserID(fmt.Sprintf("load-test-g%d-op%d", goroutineID, op))
							email := fmt.Sprintf("load%d-%d@example.com", goroutineID, op)
							name := fmt.Sprintf("Load Test User %d-%d", goroutineID, op)

							_, err := userService.CreateUser(ctx, id, email, name)
							results <- err
						}
					}(g)
				}

				wg.Wait()
				close(results)

				duration := time.Since(startTime)

				// Verify all operations completed
				totalOps := 0
				successCount := 0
				for err := range results {
					totalOps++
					if err == nil {
						successCount++
					}
				}

				Expect(totalOps).To(Equal(numGoroutines * operationsPerGoroutine))
				Expect(successCount).To(Equal(totalOps)) // All should succeed since unique emails

				// Performance should be reasonable (this is a basic smoke test)
				operationsPerSecond := float64(totalOps) / duration.Seconds()
				Expect(operationsPerSecond).To(BeNumerically(">", 100), "should handle at least 100 ops/sec")

				By(fmt.Sprintf("Completed %d operations in %v (%.2f ops/sec)",
					totalOps, duration, operationsPerSecond))
			})
		})
	})
})
