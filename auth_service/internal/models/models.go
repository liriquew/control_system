package models

type User struct {
	UID      int64
	Username string
	PassHash []byte
}
