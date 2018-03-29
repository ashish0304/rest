package main

import (
  "fmt"
  "strconv"
  "net/http"
  "github.com/gin-gonic/gin"
)

type Item struct {
  Id int `db:"id" json:"id" `
  Description string `db:"description" json:"description" `
  Hsn NullString `db:"hsn" json:"hsn" `
  Cost float32 `db:"cost" json:"cost" `
  Price float32 `db:"price" json:"price" `
  Tax float32 `db:"tax" json:"tax" `
}

type ItemTran struct{
  Type string `db:"type" json:"type"`
  Date string `db:"date" json:"date"`
  Quantity int32 `db:"quantity" json:"quantity"`
  Rate float32 `db:"rate" json:"rate"`
}

func items(c *gin.Context) {
  items:=[]Item{}
  err := DB.Select(&items, "select * from item")
  if err == nil{
    c.JSON(200, items)
  }else{
    c.JSON(404, err)
  }
}

func itemsdesc(c *gin.Context) {
  desc := "%"+c.Param("desc")+"%"
  items := []Item{}
  err := DB.Select(&items, "select * from item where description like ?", desc)
  if err == nil{
    c.JSON(200, items)
  }else{
    c.JSON(404, err)
  }
}

func itemid(c *gin.Context) {
  id := c.Param("id")
  item := Item{}
  err := DB.Get(&item, "select * from item where id=?", id)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, item)
  }
}

func itemadd(c *gin.Context) {
  item := Item{}
  if err := c.BindJSON(&item); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err})
    fmt.Printf("%#v \n%#v", item, err)
    return
  }
  if len(item.Description) < 5 || item.Cost < 0 || item.Price < 0 || item.Tax < 0{
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid values!"})
    return
  }
  _, err := DB.NamedExec("insert into item(description, hsn, cost, price, tax) values(:description, :hsn, :cost, :price, :tax)", &item)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, item)
  }
}

func itemupdate(c *gin.Context) {
  item := Item{}
  if err := c.BindJSON(&item); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err})
    fmt.Printf("%#v \n%#v", item, err)
    return
  }
  if len(item.Description) < 5 || item.Cost < 0 || item.Price < 0 || item.Tax < 0{
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid values!"})
    return
  }
  _, err := DB.NamedExec("update item set description=:description, hsn=:hsn, cost=:cost, price=:price, tax=:tax where id=:id", &item)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, item)
  }
}

func itemtran(c *gin.Context) {
  id := c.Param("id")
  offset, e1 := strconv.Atoi(c.Request.URL.Query().Get("offset"))
  if e1 != nil { offset = 0}
  limit, e2 := strconv.Atoi(c.Request.URL.Query().Get("limit"))
  if e2 != nil { limit = 10}
  trns := []ItemTran{}
  err := DB.Select(&trns, `select type, strftime('%d-%m-%Y', date) as date,
         quantity, rate from stktran where itm_id=? order by strftime('%Y-%m-%d', date) desc limit ? offset ?`, id, limit, offset)
  if err != nil {
    c.JSON(400, err)
    //fmt.Println(err)
  }else{
    c.JSON(200, trns)
  }
}
