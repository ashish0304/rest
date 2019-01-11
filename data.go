package main

import (
	"context"
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
