package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Item struct {
	Id          int        `db:"id" json:"id" `
	Description string     `db:"description" json:"description" `
	Hsn         NullString `db:"hsn" json:"hsn" `
	Cost        float32    `db:"cost" json:"cost" `
	Price       float32    `db:"price" json:"price" `
	Tax         float32    `db:"tax" json:"tax" `
}

func items(c *gin.Context) {
	desc := "%" + c.Request.URL.Query().Get("desc") + "%"
	items := []Item{}
	err := DB.Select(&items, "select * from item where description like ?", desc)
	if err == nil {
		c.JSON(http.StatusOK, items)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func itemid(c *gin.Context) {
	id := c.Param("id")
	item := Item{}
	err := DB.Get(&item, "select * from item where id=?", id)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, item)
	}
}

func itemadd(c *gin.Context) {
	item := Item{}
	if err := c.BindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, err)
		fmt.Printf("%#v \n%#v", item, err)
		return
	}
	if len(item.Description) < 5 || item.Cost < 0 || item.Price < 0 || item.Tax < 0 {
		c.JSON(http.StatusBadRequest, "Error: Invalid values!")
		return
	}
	_, err := DB.NamedExec("insert into item(description, hsn, cost, price, tax) values(:description, :hsn, :cost, :price, :tax)", &item)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, item)
	}
}

func itemupdate(c *gin.Context) {
	item := Item{}
	if err := c.BindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, err)
		fmt.Printf("%#v \n%#v", item, err)
		return
	}
	if len(item.Description) < 5 || item.Cost < 0 || item.Price < 0 || item.Tax < 0 {
		c.JSON(http.StatusBadRequest, "Error: Invalid values!")
		return
	}
	_, err := DB.NamedExec("update item set description=:description, hsn=:hsn, cost=:cost, price=:price, tax=:tax where id=:id", &item)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, item)
	}
}
