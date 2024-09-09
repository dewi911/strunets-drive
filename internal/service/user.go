package service

type User struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
}

func NewUsers(userRepo UserRepository, sessionRepo SessionRepository) *User {
	return &User{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}
