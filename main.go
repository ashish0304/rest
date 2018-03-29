package main

import (
  "github.com/jmoiron/sqlx"
  _"github.com/mattn/go-sqlite3"
  "github.com/gin-gonic/gin"
)

var Router *gin.Engine
var DB *sqlx.DB

func main() {
  DB = sqlx.MustConnect("sqlite3", "../myshop.db3")
  defer DB.Close()
  DB.Exec("PRAGMA foreign_keys = ON;")
  Router = gin.Default()

  authRead := Router.Group("/api")
  authRead.Use(AuthRead)

  authWrite := Router.Group("/api")
  authWrite.Use(AuthWrite)

  //Router.Static("/static", "../myshop/dist/static")
  authRead.GET("/items", items)
  authRead.GET("/items/:desc", itemsdesc)
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

  authWrite.POST("/stktran", stktran)
  authRead.GET("/repdateitm", repdateitm)
  authRead.GET("/replcnstat", replcnstat)
  
  authWrite.POST("/pmttran", pmttran)
  authRead.GET("/acctrans", acctrans)
  authRead.GET("/payments", payments)

  Router.POST("/api/login", login)
  Router.GET("/api/logout", logout)
  //Router.StaticFile("/", "../myshop/dist/index.html")
  Router.Run("0.0.0.0:8181")
}
