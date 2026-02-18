package services

import (
	"errors"
	"time"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      ports.UserRepository
	jwtSecret string
}

func NewUserService(repo ports.UserRepository, jwtSecret string) ports.UserUseCase {
	if jwtSecret == "" {
		jwtSecret = "fiapx-secret-key" // Default secret for dev
	}
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *userService) Register(email, password, name string) (domain.AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.repo.GetByEmail(email)
	if existingUser != nil {
		return domain.AuthResponse{}, errors.New("usuário já cadastrado com este e-mail")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.AuthResponse{}, err
	}

	user := &domain.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
	}

	err = s.repo.Create(user)
	if err != nil {
		return domain.AuthResponse{}, err
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return domain.AuthResponse{}, err
	}

	// Hide password in response
	user.Password = ""

	return domain.AuthResponse{
		User:  *user,
		Token: token,
	}, nil
}

func (s *userService) Login(email, password string) (domain.AuthResponse, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil || user == nil {
		return domain.AuthResponse{}, errors.New("credenciais inválidas")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.AuthResponse{}, errors.New("credenciais inválidas")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return domain.AuthResponse{}, err
	}

	user.Password = ""

	return domain.AuthResponse{
		User:  *user,
		Token: token,
	}, nil
}

func (s *userService) generateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
