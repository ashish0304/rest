package main

import (
	"github.com/gin-gonic/gin"
)

type Account struct {
	Id          int     `db:"id" json:"id" `
	Description string  `db:"description" json:"description" `
	Balance     float32 `db:"balance" json:"balance" `
}

func account(c *gin.Context) {
	accounts := []Account{}
	err := DB.Select(&accounts, "select * from account")
	if err == nil {
		c.JSON(200, accounts)
	} else {
		c.JSON(404, err)
	}
}

func accountid(c *gin.Context) {
	id := c.Param("id")
	account := Account{}
	err := DB.Get(&account, "select * from account where id=?", id)
	if err != nil {
		c.JSON(400, err)
	} else {
		c.JSON(200, account)
	}
}

func accountadd(c *gin.Context) {
	account := Account{}
	c.BindJSON(&account)
	_, err := DB.NamedExec("insert into account(description, balance) values(:description, :balance)", &account)
	if err != nil {
		c.JSON(400, err)
	} else {
		c.JSON(200, account)
	}
}

func accountupdate(c *gin.Context) {
	account := Account{}
	c.BindJSON(&account)
	_, err := DB.NamedExec("update account set description=:description, balance=:balance where id=:id", &account)
	if err != nil {
		c.JSON(400, err)
	} else {
		c.JSON(200, account)
	}
}
