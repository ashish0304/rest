package main

import (
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var keySecret string
var Router *gin.Engine
var DB *sqlx.DB

func main() {
	keySecret = os.Getenv("SERVERKEY")
	port := os.Getenv("SERVERPORT")
	if len(keySecret) < 2 || len(port) < 4 {
		fmt.Println("ERROR: Environment variable SERVERKEY or SERVERPORT is not defined")
		os.Exit(1)
	}

	DB = sqlx.MustConnect("sqlite3", "/sdcard/myshop/myshop.db3")
	defer DB.Close()
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
	authSales.GET("/partyacc/:desc", partyacc)
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

 authAdmin.PUT("/party", partyupdate)
 authAdmin.PUT("/rawstran", pstran)
	authAdmin.GET("/rawstran", gstran)
	authAdmin.PUT("/rawptran", pptran)
	authAdmin.GET("/rawptran", gptran)
 authAdmin.PUT("/clrstk", clrstk)
	authAdmin.PUT("/expinv", expinv)
	authAdmin.PUT("/clrinv", clrinv)
 authAdmin.POST("/location", locationadd)
	authAdmin.PUT("/location", locationupdate)
 authAdmin.POST("/account", accountadd)
	authAdmin.PUT("/account", accountupdate)
 authAdmin.PUT("/chequehonor", chequehonor)
	authAdmin.PUT("/chequecancel", chequecancel)

	Router.POST("/api/login", login)
	Router.GET("/api/logout", logout)

	Router.Run("0.0.0.0:" + port)
}
