package main

import (
	_ "github.com/dgrijalva/jwt-go"
	_ "github.com/gin-gonic/gin"
)

type User struct {
	Id          string `db:"id" json:"id"`
	Description string `db:"description" json:"description"`
	Password    string `db:"password" json:"password"`
	Access      uint32 `db:"access" json:"access"`
}
