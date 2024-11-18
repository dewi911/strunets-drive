package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"strings"
	"strunetsdrive/internal/models"
	"time"
)

type User struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
	tokenTtl    time.Duration
	jwtSecret   []byte
}

func NewUsers(userRepo UserRepository, sessionRepo SessionRepository, tokenTtl time.Duration, jwtSecret string) *User {
	return &User{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokenTtl:    tokenTtl,
		jwtSecret:   []byte(jwtSecret),
	}
}

func (s *User) SingUp(ctx context.Context, input models.SignUpInput) error {
	exist, err := s.userRepo.Exist(ctx, input.Username)
	if err != nil {
		return errors.Wrap(err, "user existence check error")
	}

	if exist {
		return errors.New("user already exists error")
	}

	//password := encrypt.Encrypt(input.Password) //TODO password err adding

	user := models.User{
		Username: input.Username,
		Password: input.Password,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return errors.Wrap(err, "create user error")
	}

	return nil
}

func (s *User) Login(ctx context.Context, input models.LoginInput) (string, string, error) {
	//password := encrypt.Encrypt(input.Password)
	//TODO password if err and pass hasher

	user, err := s.userRepo.GetByCredentials(ctx, input.Username, input.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.Wrap(err, "getting user by credential error") //TODO add constan err
		}
		return "", "", errors.Wrap(err, "getting user by credential error")
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return "", "", errors.Wrap(err, "generating tokens error")
	}

	return accessToken, refreshToken, nil
}

func (s *User) RefreshSession(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := s.sessionRepo.GetToken(ctx, refreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, "getting refresh session error")
	}

	if session.ExpiresAt.Unix() < time.Now().Unix() {
		return "", "", errors.Wrap(err, "refresh token expired error")
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return "", "", errors.Wrap(err, "getting user by id error")
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return "", "", errors.Wrap(err, "generating tokens error")
	}

	return accessToken, refreshToken, nil
}

func (s *User) ParseToken(_ context.Context, token string) (int, string, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrapf(nil, "unexpecting signing method %v", token.Header["alg"])
		}
		return s.jwtSecret, nil //TODO ADD SECRET CONST
	})

	if err != nil {
		return 0, "0", errors.Wrap(err, "jwt parsing error")
	}

	if !t.Valid {
		return 0, "0", errors.Wrap(nil, "invalid token error")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "0", errors.Wrap(nil, "invalid claims error")
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		return 0, "0", errors.Wrap(nil, "invalid subject error")
	}

	subjectParts := strings.Split(subject, ":")
	if len(subjectParts) != 2 {
		return 0, "0", errors.Wrap(nil, "token subject content error")
	}

	userID, err := strconv.Atoi(subjectParts[0])
	if err != nil {
		return 0, "0", errors.Wrap(err, "invalid userID in from token subject")
	}

	//username, err := strconv.Itoa(subjectParts[1])
	//if err != nil {
	//	return 0, "0", errors.Wrap(err, "invalid userID in from token subject error")
	//}

	return userID, subjectParts[1], nil
}

func (s *User) generateTokens(ctx context.Context, user models.User) (string, string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d:%s", user.ID, user.Username),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTtl)),
	})

	accessToken, err := t.SignedString(s.jwtSecret) //TODO add salt const
	if err != nil {
		return "", "", errors.Wrap(err, "creating and returning a complete, signed JWT token error")
	}

	refreshToken, err := newRefreshToken()
	if err != nil {
		return "", "", errors.Wrap(err, "creating new refresh token error")
	}

	if err := s.sessionRepo.Create(ctx, models.RefreshSession{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}); err != nil {
		return "", "", errors.Wrap(err, "creating refresh session error")
	}

	return accessToken, refreshToken, nil
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", errors.Wrap(err, "read generates random bytes error")
	}

	return fmt.Sprintf("%x", b), nil
}
