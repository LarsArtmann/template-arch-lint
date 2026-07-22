package handlers_test

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUserQueryHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserQueryHandler Suite")
}

var _ = Describe("UserQueryHandler", func() {
	var (
		userQueryService services.UserQueryService
		userQueryHandler *handlers.UserQueryHandler
		mux              *http.ServeMux
		userRepo         repositories.UserRepository
		userService      *services.UserService
	)

	BeforeEach(func() {
		mux = http.NewServeMux()

		userRepo = repositories.NewInMemoryUserRepository()
		userQueryService = services.NewUserQueryService(userRepo)
		userService = services.NewUserService(userRepo)
		userQueryHandler = handlers.NewUserQueryHandler(userQueryService)

		userQueryHandler.RegisterRoutes(mux)
	})

	createTestUser := func(email, name string) string {
		userID, err := values.GenerateUserID()
		Expect(err).ToNot(HaveOccurred())

		user, err := userService.CreateUser(context.Background(), userID, email, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(user).ToNot(BeNil())

		return userID.String()
	}

	expectEmptyArrayResponse := func(urlPath string) {
		req := httptest.NewRequest(http.MethodGet, urlPath, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		var response map[string]any

		err := json.Unmarshal(w.Body.Bytes(), &response)
		Expect(err).ToNot(HaveOccurred())
		Expect(response).To(HaveKey("data"))

		data, ok := response["data"].([]any)
		Expect(ok).To(BeTrue())
		Expect(data).To(BeEmpty())
	}

	expectBadRequestResponse := func(urlPath string) {
		req := httptest.NewRequest(http.MethodGet, urlPath, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusBadRequest))

		var response map[string]any

		err := json.Unmarshal(w.Body.Bytes(), &response)
		Expect(err).ToNot(HaveOccurred())
		Expect(response).To(HaveKey("error"))
	}

	Describe("Repository Sharing Debug", func() {
		It("should persist data across service calls", func() {
			userID := createTestUser("debug@example.com", "Debug User")

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

				req := httptest.NewRequest(http.MethodGet, "/api/v1/users/query/"+userID, nil)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response map[string]any

				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("data"))
			})
		})

		Context("when user does not exist", func() {
			It("should return 404 status", func() {
				req := httptest.NewRequest(
					http.MethodGet,
					"/api/v1/users/query/non-existent-id",
					nil,
				)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusNotFound))

				var response map[string]any

				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(HaveKey("error"))
			})
		})

		Context("when user ID is invalid", func() {
			It("should return 400 status for invalid characters", func() {
				expectBadRequestResponse("/api/v1/users/query/invalid@id")
			})
		})
	})

	Describe("ListUsers", func() {
		Context("when users exist", func() {
			It("should return all users with 200 status", func() {
				createTestUser("test1@example.com", "User 1")
				createTestUser("test2@example.com", "User 2")

				req := httptest.NewRequest(http.MethodGet, "/api/v1/users/query", nil)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)

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
				expectEmptyArrayResponse("/api/v1/users/query")
			})
		})
	})

	Describe("SearchUsers", func() {
		Context("when user exists with email", func() {
			It("should return user with 200 status", func() {
				createTestUser("search@example.com", "Search User")

				req := httptest.NewRequest(
					http.MethodGet,
					"/api/v1/users/search?email=search@example.com",
					nil,
				)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)

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
				expectBadRequestResponse("/api/v1/users/search")
			})
		})

		Context("when user does not exist with email", func() {
			It("should return empty array with 200 status", func() {
				expectEmptyArrayResponse("/api/v1/users/search?email=nonexistent@example.com")
			})
		})
	})

	Describe("GetUsersWithPagination", func() {
		Context("with valid pagination parameters", func() {
			It("should return paginated results", func() {
				for i := 1; i <= 5; i++ {
					createTestUser("user"+strconv.Itoa(i)+"@example.com", "User "+strconv.Itoa(i))
				}

				req := httptest.NewRequest(
					http.MethodGet,
					"/api/v1/users/paginated?page=1&limit=3",
					nil,
				)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)

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
				req := httptest.NewRequest(http.MethodGet, "/api/v1/users/paginated", nil)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, req)

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
