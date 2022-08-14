package main

import (
	"database/sql"
	"fmt"
	"github.com/gchaincl/sqlhooks"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"log"
	"os"
)

var keySecret string
var Router *gin.Engine
var logSql *log.Logger

func main() {
	keySecret = os.Getenv("SERVERKEY")
	port := os.Getenv("SERVERPORT")
	if len(keySecret) < 2 || len(port) < 4 {
		fmt.Println("ERROR: Environment variable SERVERKEY or SERVERPORT is not defined")
		os.Exit(1)
	}

	sql.Register("sqlite3WithHooks", sqlhooks.Wrap(&sqlite3.SQLiteDriver{}, &Hooks{}))

	DB = sqlx.MustConnect("sqlite3WithHooks", "../myshop.db3")
	defer DB.Close()

	f, err := os.OpenFile(os.Getenv("HOME")+"/myshop.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	logSql = log.New(f, "", log.LstdFlags)
	DB.Exec("PRAGMA foreign_keys = ON;")

	Router = gin.Default()

	authSales := Router.Group("/api")
	authSales.Use(AuthSales)

	authAdmin := Router.Group("/api")
	authAdmin.Use(AuthAdmin)

	Router.Use(static.Serve("/", static.LocalFile("./static", true)))

	authSales.GET("/items", items)
	authSales.GET("/item/:id", itemid)
	authSales.GET("/item/:id/trans", itemtran)
	authSales.POST("/item", itemadd)
	authSales.PUT("/item", itemupdate)
	authSales.GET("/parties", parties)
	authSales.GET("/parties/:desc", partiesdesc)
	authSales.GET("/party/:id", partyid)
	authSales.POST("/party", partyadd)
	authSales.GET("/party/:id/payments", prtpayments)
	authSales.GET("/party/:id/items", prtitems)
	authSales.GET("/party/:id/summary", prtsumry)
	authSales.GET("/partyacc/:desc", partyacc)
	authSales.GET("/partystk/:id", partystk)
	authSales.GET("/partiesbal", partiesbal)
	authSales.GET("/cheques", cheques)
	authSales.GET("/account", account)
	authSales.GET("/account/:id", accountid)
	authSales.GET("/location", location)
	authSales.GET("/location/:id", locationid)
	authSales.GET("/stock/:lc/:it", stock)
	authSales.GET("/stocks/:lc/:ta", stocks)
	authSales.GET("/inventory", invs)
	authSales.POST("/inventory", invsadd)
	authSales.PUT("/inventory", invsdel)
	authSales.POST("/stktran", stktran)
	authSales.GET("/repdateitm", repdateitm)
	authSales.GET("/replcnstat", replcnstat)
	authSales.POST("/pmttran", pmttran)
	authSales.GET("/acctrans", acctrans)
	authSales.GET("/payments", payments)
    authSales.GET("/gstrep", gstreps)

	authAdmin.PUT("/party", partyupdate)
	authAdmin.PUT("/clrstk", clrstk)
	authAdmin.PUT("/expinv", expinv)
	authAdmin.PUT("/clrinv", clrinv)
	authAdmin.POST("/location", locationadd)
	authAdmin.PUT("/location", locationupdate)
	authAdmin.POST("/account", accountadd)
	authAdmin.PUT("/account", accountupdate)
	authAdmin.PUT("/chequehonor", chequehonor)
	authAdmin.PUT("/chequecancel", chequecancel)
	authAdmin.GET("/log", dumplog)

	Router.POST("/api/login", login)
	Router.GET("/api/logout", logout)

	Router.Run("0.0.0.0:" + port)
}
