package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

// mockRepositoryForBench is a simple in-memory repository for benchmarking.
type mockRepositoryForBench struct {
	users map[string]*entities.User
}

func newMockRepositoryForBench() *mockRepositoryForBench {
	return &mockRepositoryForBench{
		users: make(map[string]*entities.User),
	}
}

func (m *mockRepositoryForBench) Save(_ context.Context, user *entities.User) error {
	m.users[user.ID.String()] = user

	return nil
}

func (m *mockRepositoryForBench) FindByID(_ context.Context, id values.UserID) (*entities.User, error) {
	user, exists := m.users[id.String()]
	if !exists {
		return nil, repositories.ErrUserNotFound
	}

	return user, nil
}

func (m *mockRepositoryForBench) FindByEmail(_ context.Context, email string) (*entities.User, error) {
	for _, user := range m.users {
		if user.GetEmail().String() == email {
			return user, nil
		}
	}

	return nil, repositories.ErrUserNotFound
}

func (m *mockRepositoryForBench) FindByUsername(_ context.Context, username string) (*entities.User, error) {
	for _, user := range m.users {
		if user.GetUserName().String() == username {
			return user, nil
		}
	}

	return nil, repositories.ErrUserNotFound
}

func (m *mockRepositoryForBench) List(_ context.Context) ([]*entities.User, error) {
	users := make([]*entities.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}

	return users, nil
}

func (m *mockRepositoryForBench) Delete(_ context.Context, id values.UserID) error {
	delete(m.users, id.String())

	return nil
}

// setupBenchmarkService creates a user service with pre-populated data for benchmarks.
func setupBenchmarkService(b *testing.B, userCount int) (*UserService, context.Context) {
	b.Helper()

	repo := newMockRepositoryForBench()
	service := NewUserService(repo)
	ctx := context.Background()

	// Pre-populate with test users for realistic benchmarks
	for i := range userCount {
		userID, _ := values.NewUserID(fmt.Sprintf("user-%d", i))
		email := fmt.Sprintf("user%d@example.com", i)
		name := fmt.Sprintf("User %d", i)

		user, err := entities.NewUser(userID, email, name)
		if err != nil {
			b.Fatalf("Failed to create test user: %v", err)
		}

		err = repo.Save(ctx, user)
		if err != nil {
			b.Fatalf("Failed to save test user: %v", err)
		}
	}

	return service, ctx
}

// BenchmarkCreateUser measures user creation performance.
func BenchmarkCreateUser(b *testing.B) {
	service, ctx := setupBenchmarkService(b, 0)

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("bench-user-%d", i))
		email := fmt.Sprintf("benchuser%d@example.com", i)
		name := fmt.Sprintf("Bench User %d", i)

		_, err := service.CreateUser(ctx, userID, email, name)
		if err != nil {
			b.Fatalf("CreateUser failed: %v", err)
		}
	}
}

// BenchmarkGetUser measures user retrieval performance.
func BenchmarkGetUser(b *testing.B) {
	const userCount = 1000
	service, ctx := setupBenchmarkService(b, userCount)

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("user-%d", i%userCount))

		_, err := service.GetUser(ctx, userID)
		if err != nil {
			b.Fatalf("GetUser failed: %v", err)
		}
	}
}

// BenchmarkListUsers measures user listing performance with different dataset sizes.
func BenchmarkListUsers(b *testing.B) {
	testCases := []struct {
		name      string
		userCount int
	}{
		{"100Users", 100},
		{"1000Users", 1000},
		{"10000Users", 10000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			service, ctx := setupBenchmarkService(b, tc.userCount)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				_, err := service.ListUsers(ctx)
				if err != nil {
					b.Fatalf("ListUsers failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkUpdateUser measures user update performance.
func BenchmarkUpdateUser(b *testing.B) {
	const userCount = 1000
	service, ctx := setupBenchmarkService(b, userCount)

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("user-%d", i%userCount))
		newEmail := fmt.Sprintf("updated%d@example.com", i)
		newName := fmt.Sprintf("Updated User %d", i)

		_, err := service.UpdateUser(ctx, userID, newEmail, newName)
		if err != nil {
			b.Fatalf("UpdateUser failed: %v", err)
		}
	}
}

// BenchmarkDeleteUser measures user deletion performance.
func BenchmarkDeleteUser(b *testing.B) {
	// Create fresh users for each benchmark run
	service, ctx := setupBenchmarkService(b, b.N)

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("user-%d", i))

		err := service.DeleteUser(ctx, userID)
		if err != nil {
			b.Fatalf("DeleteUser failed: %v", err)
		}
	}
}

// BenchmarkFilterActiveUsers measures functional filtering performance.
func BenchmarkFilterActiveUsers(b *testing.B) {
	const userCount = 10000
	service, ctx := setupBenchmarkService(b, userCount)

	b.ReportAllocs()

	for b.Loop() {
		_, err := service.FilterActiveUsers(ctx)
		if err != nil {
			b.Fatalf("FilterActiveUsers failed: %v", err)
		}
	}
}

// BenchmarkGetUserStats measures statistics calculation performance.
func BenchmarkGetUserStats(b *testing.B) {
	testCases := []struct {
		name      string
		userCount int
	}{
		{"1000Users", 1000},
		{"10000Users", 10000},
		{"50000Users", 50000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			service, ctx := setupBenchmarkService(b, tc.userCount)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				_, err := service.GetUserStats(ctx)
				if err != nil {
					b.Fatalf("GetUserStats failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkGetUsersByEmailDomains measures domain filtering performance.
func BenchmarkGetUsersByEmailDomains(b *testing.B) {
	const userCount = 10000
	service, ctx := setupBenchmarkService(b, userCount)

	domains := []string{"example.com", "test.com", "demo.com"}

	b.ReportAllocs()

	for b.Loop() {
		_, err := service.GetUsersByEmailDomains(ctx, domains)
		if err != nil {
			b.Fatalf("GetUsersByEmailDomains failed: %v", err)
		}
	}
}

// BenchmarkConcurrentOperations measures performance under concurrent load.
func BenchmarkConcurrentOperations(b *testing.B) {
	const userCount = 1000
	service, ctx := setupBenchmarkService(b, userCount)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Mix of operations to simulate realistic usage
			switch i % 4 {
			case 0:
				// Read operation (most common)
				userID, _ := values.NewUserID(fmt.Sprintf("user-%d", i%userCount))
				_, _ = service.GetUser(ctx, userID)
			case 1:
				// List operation
				_, _ = service.ListUsers(ctx)
			case 2:
				// Update operation
				userID, _ := values.NewUserID(fmt.Sprintf("user-%d", i%userCount))
				email := fmt.Sprintf("concurrent%d@example.com", i)
				name := fmt.Sprintf("Concurrent User %d", i)
				_, _ = service.UpdateUser(ctx, userID, email, name)
			case 3:
				// Stats operation
				_, _ = service.GetUserStats(ctx)
			}
			i++
		}
	})
}

// BenchmarkMemoryAllocation specifically measures memory allocation patterns.
func BenchmarkMemoryAllocation(b *testing.B) {
	service, ctx := setupBenchmarkService(b, 0)

	b.ReportAllocs()

	// Force memory allocations to measure GC impact
	for i := 0; b.Loop(); i++ {
		// Create temporary data structures
		userID, _ := values.NewUserID(fmt.Sprintf("memory-test-%d", i))
		email := fmt.Sprintf("memtest%d@example.com", i)
		name := fmt.Sprintf("Memory Test User %d", i)

		user, err := service.CreateUser(ctx, userID, email, name)
		if err != nil {
			b.Fatalf("CreateUser failed: %v", err)
		}

		// Immediately access the user to ensure it's not optimized away
		_ = user.ID.String()
		_ = user.GetEmail().String()
		_ = user.GetUserName().String()
	}
}

// BenchmarkValueObjectCreation measures the performance of value object creation.
func BenchmarkValueObjectCreation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		userID, err := values.NewUserID(fmt.Sprintf("value-test-%d", i))
		if err != nil {
			b.Fatalf("NewUserID failed: %v", err)
		}

		// Access the value to ensure it's not optimized away
		_ = userID.String()
	}
}

// BenchmarkEntityCreation measures the performance of entity creation.
func BenchmarkEntityCreation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		userID, _ := values.NewUserID(fmt.Sprintf("entity-test-%d", i))
		email := fmt.Sprintf("entitytest%d@example.com", i)
		name := fmt.Sprintf("Entity Test User %d", i)

		user, err := entities.NewUser(userID, email, name)
		if err != nil {
			b.Fatalf("NewUser failed: %v", err)
		}

		// Access fields to ensure the entity is not optimized away
		_ = user.ID.String()
		_ = user.GetEmail().String()
		_ = user.GetUserName().String()
		_ = user.Created
	}
}
