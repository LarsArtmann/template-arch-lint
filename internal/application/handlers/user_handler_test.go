package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

// MockUserRepository implements repositories.UserRepository for testing
type MockUserRepository struct {
	users       map[string]*entities.User
	shouldError bool
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:       make(map[string]*entities.User),
		shouldError: false,
	}
}

func (m *MockUserRepository) Save(ctx context.Context, user *entities.User) error {
	if m.shouldError {
		return fmt.Errorf("repository error")
	}
	m.users[user.ID.String()] = user
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id values.UserID) (*entities.User, error) {
	if m.shouldError {
		return nil, repositories.ErrUserNotFound
	}

	user, exists := m.users[id.String()]
	if !exists {
		return nil, repositories.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.shouldError {
		return nil, repositories.ErrUserNotFound
	}

	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, repositories.ErrUserNotFound
}

func (m *MockUserRepository) List(ctx context.Context) ([]*entities.User, error) {
	if m.shouldError {
		return nil, fmt.Errorf("repository error")
	}

	users := make([]*entities.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id values.UserID) error {
	if m.shouldError {
		return repositories.ErrUserNotFound
	}

	_, exists := m.users[id.String()]
	if !exists {
		return repositories.ErrUserNotFound
	}

	delete(m.users, id.String())
	return nil
}

func (m *MockUserRepository) SetShouldError(shouldError bool) {
	m.shouldError = shouldError
}

var _ = Describe("UserHandler", func() {
	var (
		handler     *UserHandler
		mockRepo    *MockUserRepository
		userService *services.UserService
		router      *gin.Engine
		logger      *slog.Logger
	)

	BeforeEach(func() {
		// Set Gin to test mode
		gin.SetMode(gin.TestMode)

		// Create mock repository and logger
		mockRepo = NewMockUserRepository()
		logger = slog.New(slog.NewTextHandler(GinkgoWriter, nil))

		// Create service with mock repository
		userService = services.NewUserService(mockRepo)

		// Create handler
		handler = NewUserHandler(userService, logger)

		// Setup router
		router = gin.New()
		router.POST("/users", handler.CreateUser)
		router.GET("/users/:id", handler.GetUser)
		router.PUT("/users/:id", handler.UpdateUser)
		router.DELETE("/users/:id", handler.DeleteUser)
		router.GET("/users", handler.ListUsers)
		router.GET("/users-stats", handler.GetUserStats)
		router.GET("/active-users", handler.GetActiveUsers)
		router.GET("/user-emails", handler.GetUserEmails)
	})

	Describe("CreateUser", func() {
		Context("with valid request", func() {
			It("should create user successfully", func() {
				// Given
				requestBody := CreateUserRequest{
					ID:    "user-123",
					Email: "test@example.com",
					Name:  "TestUser",
				}
				jsonBody, err := json.Marshal(requestBody)
				Expect(err).To(BeNil())

				req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusCreated))

				var response entities.User
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).To(BeNil())
				Expect(response.ID.String()).To(Equal("user-123"))
				Expect(response.Email).To(Equal("test@example.com"))
				Expect(response.Name).To(Equal("TestUser"))
			})
		})

		Context("with invalid request", func() {
			It("should return error for malformed JSON", func() {
				// Given
				invalidJSON := `{"id": "user-123", "email": "test@example.com", "name": "TestUser"`
				req := httptest.NewRequest("POST", "/users", bytes.NewBuffer([]byte(invalidJSON)))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).To(BeNil())
				Expect(response["error"]).To(Equal("Invalid request payload"))
			})

			It("should return error for missing required fields", func() {
				// Given
				requestBody := CreateUserRequest{
					// Missing ID
					Email: "test@example.com",
					Name:  "TestUser",
				}
				jsonBody, err := json.Marshal(requestBody)
				Expect(err).To(BeNil())

				req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("should return error for invalid email format", func() {
				// Given
				requestBody := CreateUserRequest{
					ID:    "user-123",
					Email: "invalid-email", // Invalid email
					Name:  "TestUser",
				}
				jsonBody, err := json.Marshal(requestBody)
				Expect(err).To(BeNil())

				req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})

			It("should return error for invalid user ID format", func() {
				// Given
				requestBody := CreateUserRequest{
					ID:    "invalid id with spaces", // Invalid user ID
					Email: "test@example.com",
					Name:  "TestUser",
				}
				jsonBody, err := json.Marshal(requestBody)
				Expect(err).To(BeNil())

				req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).To(BeNil())
				Expect(response["error"]).To(Equal("Invalid user ID format"))
			})
		})
	})

	Describe("GetUser", func() {
		BeforeEach(func() {
			// Create a test user
			userID, _ := values.NewUserID("user-123")
			user, _ := entities.NewUser(userID, "test@example.com", "TestUser")
			mockRepo.users["user-123"] = user
		})

		Context("with valid user ID", func() {
			It("should return user successfully", func() {
				// Given
				req := httptest.NewRequest("GET", "/users/user-123", nil)
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response entities.User
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).To(BeNil())
				Expect(response.ID.String()).To(Equal("user-123"))
			})
		})

		Context("with invalid user ID", func() {
			It("should return error for invalid ID format", func() {
				// Given
				req := httptest.NewRequest("GET", "/users/invalid%20id", nil)
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("with non-existent user", func() {
			It("should return not found error", func() {
				// Given
				req := httptest.NewRequest("GET", "/users/nonexistent", nil)
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("UpdateUser", func() {
		BeforeEach(func() {
			// Create a test user
			userID, _ := values.NewUserID("user-123")
			user, _ := entities.NewUser(userID, "test@example.com", "TestUser")
			mockRepo.users["user-123"] = user
		})

		Context("with valid request", func() {
			It("should update user successfully", func() {
				// Given
				requestBody := UpdateUserRequest{
					Email: "updated@example.com",
					Name:  "UpdatedUser",
				}
				jsonBody, err := json.Marshal(requestBody)
				Expect(err).To(BeNil())

				req := httptest.NewRequest("PUT", "/users/user-123", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response entities.User
				err = json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).To(BeNil())
				Expect(response.Email).To(Equal("updated@example.com"))
				Expect(response.Name).To(Equal("UpdatedUser"))
			})
		})
	})

	Describe("DeleteUser", func() {
		BeforeEach(func() {
			// Create a test user
			userID, _ := values.NewUserID("user-123")
			user, _ := entities.NewUser(userID, "test@example.com", "TestUser")
			mockRepo.users["user-123"] = user
		})

		Context("with valid user ID", func() {
			It("should delete user successfully", func() {
				// Given
				req := httptest.NewRequest("DELETE", "/users/user-123", nil)
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).To(BeNil())
				Expect(response["message"]).To(Equal("User deleted successfully"))
			})
		})
	})

	Describe("ListUsers", func() {
		BeforeEach(func() {
			// Create test users
			for i := 1; i <= 3; i++ {
				userID, _ := values.NewUserID(fmt.Sprintf("user-%d", i))
				user, _ := entities.NewUser(userID, fmt.Sprintf("user%d@example.com", i), fmt.Sprintf("User %d", i))
				mockRepo.users[userID.String()] = user
			}
		})

		Context("without query parameters", func() {
			It("should return all users", func() {
				// Given
				req := httptest.NewRequest("GET", "/users", nil)
				recorder := httptest.NewRecorder()

				// When
				router.ServeHTTP(recorder, req)

				// Then
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).To(BeNil())
				Expect(response["total"]).To(Equal(float64(3)))
			})
		})
	})

	Describe("Additional endpoints", func() {
		It("should get user stats", func() {
			// Given
			req := httptest.NewRequest("GET", "/users-stats", nil)
			recorder := httptest.NewRecorder()

			// When
			router.ServeHTTP(recorder, req)

			// Then
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should get active users", func() {
			// Given
			req := httptest.NewRequest("GET", "/active-users", nil)
			recorder := httptest.NewRecorder()

			// When
			router.ServeHTTP(recorder, req)

			// Then
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should get user emails", func() {
			// Given
			req := httptest.NewRequest("GET", "/user-emails", nil)
			recorder := httptest.NewRecorder()

			// When
			router.ServeHTTP(recorder, req)

			// Then
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})
})
