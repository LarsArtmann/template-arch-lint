package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

func TestUserQueryHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserQueryHandler Suite")
}

var _ = Describe("UserQueryHandler", func() {
	var (
		userQueryService services.UserQueryService
		userQueryHandler *handlers.UserQueryHandler
		router           *gin.Engine
		userRepo         repositories.UserRepository
		userService      *services.UserService
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()

		// Ensure both services use SAME repository instance
		userRepo = repositories.NewInMemoryUserRepository()
		userQueryService = services.NewUserQueryService(userRepo)
		userService = services.NewUserService(userRepo)
		userQueryHandler = handlers.NewUserQueryHandler(userQueryService)

		// Setup routes
		router.GET("/users/:id", userQueryHandler.GetUser)
		router.GET("/users", userQueryHandler.ListUsers)
		router.GET("/users/search", userQueryHandler.SearchUsers)
		router.GET("/users/domain/:domain", userQueryHandler.GetUsersByDomain)
		router.GET("/users/stats", userQueryHandler.GetUserStats)
		router.GET("/users/active", userQueryHandler.GetActiveUsers)
		router.GET("/users/paginated", userQueryHandler.GetUsersWithPagination)
	})

	// Helper function to create test user
	createTestUser := func(email, name string) string {
		// Generate unique ID for each test user
		userID, err := values.GenerateUserID()
		Expect(err).ToNot(HaveOccurred())

		// Use SAME user service that shares repository
		user, err := userService.CreateUser(context.Background(), userID, email, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(user).ToNot(BeNil())

		return userID.String()
	}

	Describe("Repository Sharing Debug", func() {
		It("should persist data across service calls", func() {
			// This test verifies repository instance sharing
			userID := createTestUser("debug@example.com", "Debug User")

			// Verify user exists in repository
			retrievedUserID, err := values.NewUserID(userID)
			Expect(err).ToNot(HaveOccurred())

			retrievedUser, err := userQueryService.GetUser(context.Background(), retrievedUserID)
			Expect(err).ToNot(HaveOccurred())
			Expect(retrievedUser).ToNot(BeNil())

			Expect(retrievedUser.ID.String()).To(Equal(userID))
			Expect(retrievedUser.GetEmail().String()).To(Equal("debug@example.com"))
			Expect(retrievedUser.GetUserName().String()).To(Equal("Debug User"))
		})
	})

	Describe("GetUser", func() {
		Context("when user exists", func() {
			It("should return user with 200 status", func() {
				userID := createTestUser("test@example.com", "Test User")

				req, _ := http.NewRequest(http.MethodGet, "/users/"+userID, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("data"))
			})
		})

		Context("when user does not exist", func() {
			It("should return 404 status", func() {
				req, _ := http.NewRequest(http.MethodGet, "/users/non-existent-id", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusNotFound))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("error"))
			})
		})

		Context("when user ID is invalid", func() {
			It("should return 400 status for invalid characters", func() {
				req, _ := http.NewRequest(http.MethodGet, "/users/invalid@id", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("error"))
			})
		})
	})

	Describe("ListUsers", func() {
		Context("when users exist", func() {
			It("should return all users with 200 status", func() {
				createTestUser("test1@example.com", "User 1")
				createTestUser("test2@example.com", "User 2")

				req, _ := http.NewRequest(http.MethodGet, "/users", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("data"))

				data := response["data"].([]any)
				Expect(len(data)).To(BeNumerically(">=", 2))
			})
		})

		Context("when no users exist", func() {
			It("should return empty array with 200 status", func() {
				req, _ := http.NewRequest(http.MethodGet, "/users", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("data"))

				data := response["data"].([]any)
				Expect(data).To(BeEmpty())
			})
		})
	})

	Describe("SearchUsers", func() {
		Context("when user exists with email", func() {
			It("should return user with 200 status", func() {
				createTestUser("search@example.com", "Search User")

				req, _ := http.NewRequest(http.MethodGet, "/users/search?email=search@example.com", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("data"))

				data := response["data"].([]any)
				Expect(data).To(HaveLen(1))
			})
		})

		Context("when email parameter is missing", func() {
			It("should return 400 status", func() {
				req, _ := http.NewRequest(http.MethodGet, "/users/search", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("error"))
			})
		})

		Context("when user does not exist with email", func() {
			It("should return empty array with 200 status", func() {
				req, _ := http.NewRequest(http.MethodGet, "/users/search?email=nonexistent@example.com", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("data"))

				data := response["data"].([]any)
				Expect(data).To(BeEmpty())
			})
		})
	})

	Describe("GetUsersWithPagination", func() {
		Context("with valid pagination parameters", func() {
			It("should return paginated results", func() {
				// Create multiple users
				for i := 1; i <= 5; i++ {
					createTestUser("user"+strconv.Itoa(i)+"@example.com", "User "+strconv.Itoa(i))
				}

				req, _ := http.NewRequest(http.MethodGet, "/users/paginated?page=1&limit=3", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("data"))
				Expect(response).To(HaveKey("pagination"))

				data := response["data"].([]any)
				pagination := response["pagination"].(map[string]any)
				Expect(data).To(HaveLen(3))
				Expect(pagination["page"]).To(Equal(float64(1)))
				Expect(pagination["limit"]).To(Equal(float64(3)))
				Expect(pagination["total"]).To(BeNumerically(">=", 5))
			})
		})

		Context("with default pagination parameters", func() {
			It("should use default values", func() {
				req, _ := http.NewRequest(http.MethodGet, "/users/paginated", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("pagination"))

				pagination := response["pagination"].(map[string]any)
				Expect(pagination["page"]).To(Equal(float64(1)))
				Expect(pagination["limit"]).To(Equal(float64(10)))
			})
		})
	})
})
