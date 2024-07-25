package models

type User struct {
	ID       uint32
	Email    string
	PassHash []byte
}
