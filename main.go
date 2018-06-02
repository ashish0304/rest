package main

import (
  "os"
  "fmt"
  "github.com/jmoiron/sqlx"
  _"github.com/mattn/go-sqlite3"
  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/contrib/static"
)

var keySecret string
var Router *gin.Engine
var DB *sqlx.DB

func main() {
  keySecret = os.Getenv("SERVERKEY")
  port := os.Getenv("SERVERPORT")
  if len(keySecret) < 2 || len(port) < 4{
    fmt.Println("ERROR: Environment variable SERVERKEY or SERVERPORT is not defined")
    os.Exit(1)
  }

  DB = sqlx.MustConnect("sqlite3", "/sdcard/myshop/myshop.db3")
  defer DB.Close()
  DB.Exec("PRAGMA foreign_keys = ON;")
  Router = gin.Default()

  authRead := Router.Group("/api")
  authRead.Use(AuthRead)

  authWrite := Router.Group("/api")
  authWrite.Use(AuthWrite)

  Router.Use(static.Serve("/", static.LocalFile("./static", true)))

  authRead.GET("/items", items)
  authRead.GET("/item/:id", itemid)
  authRead.GET("/item/:id/trans", itemtran)
  authWrite.POST("/item", itemadd)
  authWrite.PUT("/item", itemupdate)

  authRead.GET("/parties", parties)
  authRead.GET("/parties/:desc", partiesdesc)
  authRead.GET("/party/:id", partyid)
  authWrite.POST("/party", partyadd)
  authWrite.PUT("/party", partyupdate)
  authRead.GET("/party/:id/payments", prtpayments)
  authRead.GET("/party/:id/items", prtitems)
  authRead.GET("/partyacc/:desc", partyacc)
  authRead.GET("/partiesbal", partiesbal)
  
  authRead.GET("/account", account)
  authRead.GET("/account/:id", accountid)
  authWrite.POST("/account", accountadd)
  authWrite.PUT("/account", accountupdate)
  
  authRead.GET("/location", location)
  authRead.GET("/location/:id", locationid)
  authWrite.POST("/location", locationadd)
  authWrite.PUT("/location", locationupdate)

  authRead.GET("/stock/:lc/:it", stock)
  authRead.GET("/stocks/:lc/:ta", stocks)

  authRead.GET("/inventory", invs)
  authRead.POST("/inventory", invsadd)
  authRead.PUT("/inventory", invsdel)
  authRead.PUT("/clrstk", clrstk)
  authRead.PUT("/expinv", expinv)
  authRead.PUT("/clrinv", clrinv)

  authWrite.POST("/stktran", stktran)
  authRead.GET("/repdateitm", repdateitm)
  authRead.GET("/replcnstat", replcnstat)
  
  authWrite.POST("/pmttran", pmttran)
  authRead.GET("/acctrans", acctrans)
  authRead.GET("/payments", payments)

  Router.POST("/api/login", login)
  Router.GET("/api/logout", logout)

  Router.Run("0.0.0.0:" + port)
}
