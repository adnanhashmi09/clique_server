package user

import (
	"context"
	"fmt"
	"time"

	"github.com/adnanhashmi09/clique_server/utils"
	"github.com/gocql/gocql"
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
		Name:     req.Name,
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

	signed_jwt, err := utils.Generate_JWT_Token(u.ID.String(), u.Username, u.Email)

	if err != nil {
		return nil, err
	}

	return &LoginUserRes{
		accessToken: signed_jwt,
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Name:        u.Name,
	}, nil

}

func (s *Service) AllInfo(c context.Context, user_id gocql.UUID) (*UserAllInfo, error) {

	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	requesting_user_username := ctx.Value("requesting_user_username").(string)
	requesting_user_email := ctx.Value("requesting_user_email").(string)

	dm, rooms, user, err := s.REPOSITORY.FetchAllInformation(ctx, user_id, requesting_user_username, requesting_user_email)

	if err != nil {
		return nil, err
	}

	result := &UserAllInfo{
		DirectMsgChannels: dm,
		Rooms:             rooms,
		User:              user,
	}

	return result, nil
}
