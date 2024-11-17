package models

type SignUpInput struct {
	Username string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}
