package main

import (
  "fmt"
  "net/http"
  "github.com/gin-gonic/gin"
)

type Party struct{
  Id int `db:"id" json:"id"`
  Description string `db:"description" json:"description"`
  Address string `db:"address" json:"address"`
  Gstn NullString `db:"gstn" json:"gstn"`
  Balance float32 `db:"balance" json:"balance"`
}

type PartyAcc struct{
  Id int `db:"id" json:"id"`
  Description string `db:"description" json:"description"`
}

func parties(c *gin.Context) {
  parties:=[]Party{}
  err := DB.Select(&parties, "select * from party")
  if err == nil{
    c.JSON(200, parties)
  }else{
    c.JSON(404, err)
  }
}

func partyacc(c *gin.Context) {
  desc := "%"+c.Param("desc")+"%"
  parties:=[]PartyAcc{}
  err := DB.Select(&parties, `select id*-1 as id, description
                              from account where description
                              like ? union all
                              select id, description
                              from party where description
                              like ?`, desc, desc)
  if err == nil{
    c.JSON(200, parties)
  }else{
    c.JSON(404, err)
  }
}

func partiesdesc(c *gin.Context) {
  desc := "%"+c.Param("desc")+"%"
  parties := []Party{}
  err := DB.Select(&parties, "select * from party where description like ?", desc)
  if err == nil{
    c.JSON(200, parties)
  }else{
    c.JSON(404, err)
  }
}

func partyid(c *gin.Context) {
  id := c.Param("id")
  party := Party{}
  err := DB.Get(&party, "select * from party where id=?", id)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, party)
  }
}

func partyadd(c *gin.Context) {
  party := Party{}
  if err := c.BindJSON(&party); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err})
    fmt.Printf("%#v \n%#v", party, err)
    return
  }
  if len(party.Description) < 5 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Description length is less than 5!"})
    return
  }
  _, err := DB.NamedExec("insert into party(description, address, gstn, balance) values(:description, :address, :gstn, :balance)", &party)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, party)
  }
}

func partyupdate(c *gin.Context) {
  party := Party{}
  if err := c.BindJSON(&party); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err})
    fmt.Printf("%#v \n%#v", party, err)
    return
  }
  if len(party.Description) < 5 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Description length is less than 5!"})
    return
  }
  _, err := DB.NamedExec("update party set description=:description, address=:address, gstn=:gstn, balance=:balance where id=:id", &party)
  if err!=nil{
    c.JSON(400, err)
  }else{
    c.JSON(200, party)
  }
}

func partiesbal(c *gin.Context) {
  partyR := []Party{}
  partyP := []Party{}
  errR := DB.Select(&partyR, "select id, description, balance from party where balance > 0 order by description")
  errP := DB.Select(&partyP, "select id, description, balance*-1 as balance from party where balance < 0 order by description")
  if errR == nil && errP == nil {
    c.JSON(200, gin.H{"R": partyR, "P": partyP})
  }
}
