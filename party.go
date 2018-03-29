package main

import (
  "fmt"
  "strconv"
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

type PartyPmts struct{
  Type string `db:"type" json:"type"`
  Date string `db:"date" json:"date"`
  Account string `db:"account" json:"account"`
  Amount float32 `db:"amount" json:"amount"`
  Comment NullString `db:"comment" json:"comment"`
}

type PartyItms struct{
  Type string `db:"type" json:"type"`
  Date string `db:"date" json:"date"`
  Item string `db:"item" json:"item"`
  Quantity int32 `db:"quantity" json:"quantity"`
  Rate float32 `db:"rate" json:"rate"`
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

func prtpayments(c *gin.Context) {
  id := c.Param("id")
  offset, e1 := strconv.Atoi(c.Request.URL.Query().Get("offset"))
  if e1 != nil { offset = 0}
  limit, e2 := strconv.Atoi(c.Request.URL.Query().Get("limit"))
  if e2 != nil { limit = 10}
  pmts := []PartyPmts{}
  err := DB.Select(&pmts, `select type, strftime('%d-%m-%Y', date) as date,
         account.description as account, amount, comment
         from pmttran left join account on acc_id=account.id where prt_id=? order by strftime('%Y-%m-%d', date) desc limit ? offset ?`, id, limit, offset)
  if err != nil {
    c.JSON(400, err)
    //fmt.Println(err)
  }else{
    c.JSON(200, pmts)
  }
}

func prtitems(c *gin.Context) {
  id := c.Param("id")
  offset, e1 := strconv.Atoi(c.Request.URL.Query().Get("offset"))
  if e1 != nil { offset = 0}
  limit, e2 := strconv.Atoi(c.Request.URL.Query().Get("limit"))
  if e2 != nil { limit = 10}
  itms := []PartyItms{}
  err := DB.Select(&itms, `select type, strftime('%d-%m-%Y', date) as date,
         item.description as item, quantity, rate
         from stktran left join item on itm_id=item.id where prt_id=? order by strftime('%Y-%m-%d', date) desc limit ? offset ?`, id, limit, offset)
  if err != nil {
    c.JSON(400, err)
    //fmt.Println(err)
  }else{
    c.JSON(200, itms)
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
