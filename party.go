package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Party struct {
	Id          int         `db:"id" json:"id"`
	Description string      `db:"description" json:"description"`
	Address     string      `db:"address" json:"address"`
	Gstn        NullString  `db:"gstn" json:"gstn"`
	Balance     float32     `db:"balance" json:"balance"`
	Chq_amt     NullFloat64 `db:"chq_amt" json:"chq_amt"`
}

type Cheques struct {
 Prt_id int `db:"prt_id" json:"prt_id"`
 Party string `db:"party" json:"party"`
 Description NullString `db:"description" json:"description"`
 Date int `db:"date" json:"date"`
 Amount float32 `db:"amount" json:"amount"`
}

type PartyAcc struct {
	Id          int    `db:"id" json:"id"`
	Description string `db:"description" json:"description"`
}

func parties(c *gin.Context) {
	parties := []Party{}
	err := DB.Select(&parties, "select * from party")
	if err == nil {
		c.JSON(200, parties)
	} else {
		c.JSON(404, err)
	}
}

func partyacc(c *gin.Context) {
	desc := "%" + c.Param("desc") + "%"
	parties := []PartyAcc{}
	err := DB.Select(&parties, `select id*-1 as id, description
                              from account where description
                              like ? union all
                              select id, description
                              from party where description
                              like ?`, desc, desc)
	if err == nil {
		c.JSON(200, parties)
	} else {
		c.JSON(404, err)
	}
}

func partiesdesc(c *gin.Context) {
	desc := "%" + c.Param("desc") + "%"
	parties := []Party{}
	err := DB.Select(&parties, "select * from party where description like ?", desc)
	if err == nil {
		c.JSON(200, parties)
	} else {
		c.JSON(404, err)
	}
}

func partyid(c *gin.Context) {
	id := c.Param("id")
	party := Party{}
	err := DB.Get(&party, "select * from party where id=?", id)
	if err != nil {
		c.JSON(400, err)
	} else {
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
	_, err := DB.NamedExec("insert into party(description, address, gstn, balance, chq_amt) values(:description, :address, :gstn, :balance, :chq_amt)", &party)
	if err != nil {
		c.JSON(400, err)
	} else {
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
	_, err := DB.NamedExec("update party set description=:description, address=:address, gstn=:gstn, balance=:balance, chq_amt=:chq_amt where id=:id", &party)
	if err != nil {
		c.JSON(400, err)
	} else {
		c.JSON(200, party)
	}
}

func partiesbal(c *gin.Context) {
	partyR := []Party{}
	partyP := []Party{}
	errR := DB.Select(&partyR, "select id, description, balance, chq_amt from party where balance > 0 order by description")
	errP := DB.Select(&partyP, "select id, description, balance*-1, chq_amt*-1 from party where balance < 0 order by description")
	if errR == nil && errP == nil {
		c.JSON(200, gin.H{"R": partyR, "P": partyP})
	}
}

func cheques(c *gin.Context) {
	chequeR := []Cheques{}
	chequeP := []Cheques{}
	errR := DB.Select(&chequeR, `
  select prt_id, party.description as party,
  cheque.description, date, amount from cheque
  left join party on cheque.prt_id=party.id
  where amount > 0 order by date`)
	errP := DB.Select(&chequeP, `
  select prt_id, party.description as party,
  cheque.description, date, amount from cheque
  left join party on cheque.prt_id=party.id
  where amount < 0 order by date`)
	if errR == nil && errP == nil {
		c.JSON(200, gin.H{"R": chequeR, "P": chequeP})
	}
}
