package main

import (
  _"fmt"
  "github.com/gin-gonic/gin"
)

type Location struct {
	Id int `db:"id" json:"id" `
	Description string `db:"description" json:"description" `
	Address NullString `db:"address" json:"address" `
	Dummy NullBool `db:"dummy" json:"dummy"`
	Lst_sid uint `db:"lst_sid" json:"lst_sid"`
	Lst_pid uint `db:"lst_pid" json:"lst_pid"`
	Lst_tid uint `db:"lst_tid" json:"lst_tid"`
	Lst_aid uint `db:"lst_aid" json:"lst_aid"`
}

func location(c *gin.Context) {
  locations:=[]Location{}
  err := DB.Select(&locations, "select * from location")
  //fmt.Printf("%#v", locations)
  if err == nil{
    c.JSON(200, locations)
  }else{
    c.JSON(404, err)
  }
}

func locationid(c *gin.Context) {
  id := c.Param("id")
  location := Location{}
  err := DB.Get(&location, "select * from location where id=?", id)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, location)
  }
}

func locationadd(c *gin.Context) {
  location := Location{}
  c.BindJSON(&location)
  _, err := DB.NamedExec("insert into location(description, address, dummy) values(:description, :address, :dummy)", &location)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, location)
  }
}

func locationupdate(c *gin.Context) {
  location := Location{}
  c.BindJSON(&location)
  _, err := DB.NamedExec("update location set description=:description, address=:address, dummy=:dummy where id=:id", &location)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, location)
  }
}

