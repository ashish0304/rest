package main

import (
	"context"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

type Hooks struct{}

func (h *Hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return ctx, nil
}

func (h *Hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	logSql.Println(strings.Join(strings.Fields(query), " "))
	if len(args) > 0 {
		logSql.Println(args...)
	}
	return ctx, nil
}

func backup(c *gin.Context) {
	cmd := exec.Command("bkp")
	err := cmd.Run()
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.Status(200)
	}
}

func dumplog(c *gin.Context) {
	c.Header("Content-Disposition", "attachment;filename=log.txt")
	c.Header("Content-Type", "application/octet-stream")
	c.File("../myshop.log")
}
