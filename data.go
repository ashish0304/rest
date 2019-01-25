package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"strings"
)

var DB *sqlx.DB

type Hooks struct{}

func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return ctx, nil
}

func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	logSql.Println(strings.Join(strings.Fields(query), " "))
	logSql.Println(args)
	return ctx, nil
}

func dumplog(c *gin.Context) {
	c.Header("Content-Disposition", "attachment;filename=log.txt")
	c.Header("Content-Type", "application/octet-stream")
	c.File("../myshop.log")
}
