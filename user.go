package main

const (
	UserSales uint32 = 1
	UserAdmin uint32 = 0xffffffff
)

type User struct {
	Id          string `db:"id" json:"id"`
	Description string `db:"description" json:"description"`
	Password    string `db:"password" json:"password"`
	Access      uint32 `db:"access" json:"access"`
}
