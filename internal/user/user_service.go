package user

import (
	"context"
	"fmt"
	"time"

	"github.com/adnanhashmi09/clique_server/utils"
	// "github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	REPOSITORY
	timeout time.Duration
}

func NewService(repository REPOSITORY) SERVICE {

	return &Service{
		repository,
		time.Duration(2) * time.Second,
	}
}

const (
	secretKey = "secret"
)

func (s *Service) CreateUser(c context.Context, req *CreateUserReq) (*CreateUserRes, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	hashed_password, err := utils.HashPassword(req.Password)

	if err != nil {
		return nil, err
	}

	new_user := &User{
		Username: req.Username,
		Email:    req.Email,
		Name:     req.Email,
		Password: hashed_password,
	}

	repo, err := s.REPOSITORY.CreateUser(ctx, new_user)
	if err != nil {
		return nil, err
	}

	res := &CreateUserRes{
		ID:       repo.ID,
		Username: repo.Username,
		Email:    repo.Email,
		Name:     repo.Name,
	}

	return res, nil

}

type CustomClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (s *Service) Login(c context.Context, req *LoginUserReq) (*LoginUserRes, error) {

	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	u, err := s.REPOSITORY.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	err = utils.CheckPassword(req.Password, u.Password)
	if err != nil {
		return nil, fmt.Errorf("Passwords don't match. %v", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		ID:       u.ID.String(),
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    u.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	return &LoginUserRes{
		accessToken: ss,
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Name:        u.Name,
	}, nil

}
