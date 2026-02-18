package services

import (
	"errors"
	"testing"
	"video-processor/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Register(t *testing.T) {
	jwtSecret := "test-secret"

	t.Run("success", func(t *testing.T) {
		repo := new(MockUserRepository)
		service := NewUserService(repo, jwtSecret)

		email := "test@example.com"
		password := "password123"
		name := "Test User"

		repo.On("GetByEmail", email).Return(nil, nil)
		repo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

		resp, err := service.Register(email, password, name)

		assert.NoError(t, err)
		assert.Equal(t, email, resp.User.Email)
		assert.Equal(t, name, resp.User.Name)
		assert.NotEmpty(t, resp.Token)
		assert.Empty(t, resp.User.Password)
		repo.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		repo := new(MockUserRepository)
		service := NewUserService(repo, jwtSecret)

		email := "existing@example.com"
		repo.On("GetByEmail", email).Return(&domain.User{Email: email}, nil)

		resp, err := service.Register(email, "pass", "Name")

		assert.Error(t, err)
		assert.Equal(t, "usuário já cadastrado com este e-mail", err.Error())
		assert.Empty(t, resp.Token)
		repo.AssertExpectations(t)
	})

	t.Run("repo create error", func(t *testing.T) {
		repo := new(MockUserRepository)
		service := NewUserService(repo, jwtSecret)

		email := "test@example.com"
		repo.On("GetByEmail", email).Return(nil, nil)
		repo.On("Create", mock.AnythingOfType("*domain.User")).Return(errors.New("db error"))

		_, err := service.Register(email, "password", "Name")

		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestUserService_Login(t *testing.T) {
	jwtSecret := "test-secret"

	t.Run("success", func(t *testing.T) {
		repo := new(MockUserRepository)
		service := NewUserService(repo, jwtSecret)

		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := &domain.User{
			ID:       1,
			Email:    email,
			Password: string(hashedPassword),
			Name:     "Test User",
		}

		repo.On("GetByEmail", email).Return(user, nil)

		resp, err := service.Login(email, password)

		assert.NoError(t, err)
		assert.Equal(t, email, resp.User.Email)
		assert.NotEmpty(t, resp.Token)
		assert.Empty(t, resp.User.Password)
		repo.AssertExpectations(t)
	})

	t.Run("invalid credentials - wrong password", func(t *testing.T) {
		repo := new(MockUserRepository)
		service := NewUserService(repo, jwtSecret)

		email := "test@example.com"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

		user := &domain.User{
			Email:    email,
			Password: string(hashedPassword),
		}

		repo.On("GetByEmail", email).Return(user, nil)

		_, err := service.Login(email, "wrong-password")

		assert.Error(t, err)
		assert.Equal(t, "credenciais inválidas", err.Error())
		repo.AssertExpectations(t)
	})

	t.Run("invalid credentials - user not found", func(t *testing.T) {
		repo := new(MockUserRepository)
		service := NewUserService(repo, jwtSecret)

		email := "nonexistent@example.com"
		repo.On("GetByEmail", email).Return(nil, nil)

		_, err := service.Login(email, "password")

		assert.Error(t, err)
		assert.Equal(t, "credenciais inválidas", err.Error())
		repo.AssertExpectations(t)
	})
}
