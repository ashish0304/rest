package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Party struct {
	Id          int        `db:"id" json:"id"`
	Description string     `db:"description" json:"description"`
	Address     string     `db:"address" json:"address"`
	Gstn        NullString `db:"gstn" json:"gstn"`
	Balance     float32    `db:"balance" json:"balance"`
	Chq_amt     float32    `db:"chq_amt" json:"chq_amt"`
}

type Cheques struct {
	Rid         uint       `db:"rowid" json:"rowid"`
	Acc_id      uint       `db:"acc_id" json:"acc_id"`
	Account     string     `db:"account" json:"account"`
	Prt_id      uint       `db:"prt_id" json:"prt_id"`
	Party       string     `db:"party" json:"party"`
	Description NullString `db:"description" json:"description"`
	Date        int        `db:"date" json:"date"`
	Amount      float32    `db:"amount" json:"amount"`
}

func parties(c *gin.Context) {
	parties := []Party{}
	err := DB.Select(&parties, "select * from party")
	if err == nil {
		c.JSON(http.StatusOK, parties)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func partyacc(c *gin.Context) {
	desc := "%" + c.Param("desc") + "%"
	parties := []PartyAcc{}
	err := DB.Select(&parties, `select id*-1 as id, description, balance
                              from account where description
                              like ? union all
                              select id, description, balance
                              from party where description
                              like ?`, desc, desc)
	if err == nil {
		c.JSON(http.StatusOK, parties)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func partiesdesc(c *gin.Context) {
	desc := "%" + c.Param("desc") + "%"
	parties := []Party{}
	err := DB.Select(&parties, "select * from party where description like ?", desc)
	if err == nil {
		c.JSON(http.StatusOK, parties)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func partyid(c *gin.Context) {
	id := c.Param("id")
	party := Party{}
	err := DB.Get(&party, "select * from party where id=?", id)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, party)
	}
}

func partyadd(c *gin.Context) {
	party := Party{}
	if err := c.BindJSON(&party); err != nil {
		c.JSON(http.StatusBadRequest, err)
		fmt.Printf("%#v \n%#v", party, err)
		return
	}
	if len(party.Description) < 5 {
		c.JSON(http.StatusBadRequest, "Error: Description length is less than 5!")
		return
	}
	_, err := DB.NamedExec("insert into party(description, address, gstn, balance, chq_amt) values(:description, :address, :gstn, :balance, :chq_amt)", &party)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, party)
	}
}

func partyupdate(c *gin.Context) {
	party := Party{}
	if err := c.BindJSON(&party); err != nil {
		c.JSON(http.StatusBadRequest, err)
		fmt.Printf("%#v \n%#v", party, err)
		return
	}
	if len(party.Description) < 5 {
		c.JSON(http.StatusBadRequest, "Error: Description length is less than 5!")
		return
	}
	_, err := DB.NamedExec("update party set description=:description, address=:address, gstn=:gstn, balance=:balance, chq_amt=:chq_amt where id=:id", &party)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, party)
	}
}

func partiesbal(c *gin.Context) {
	partyR := []Party{}
	partyP := []Party{}
	errR := DB.Select(&partyR, "select id, description, balance, chq_amt from party where balance > 0 order by description")
	errP := DB.Select(&partyP, "select id, description, balance*-1 as balance, chq_amt*-1 as chq_amt from party where balance < 0 order by description")
	if errR == nil && errP == nil {
		c.JSON(http.StatusOK, gin.H{"R": partyR, "P": partyP})
	}
}

func cheques(c *gin.Context) {
	chequeR := []Cheques{}
	chequeP := []Cheques{}
	errR := DB.Select(&chequeR, `
  select cheque.rowid, prt_id, party.description as party,
  acc_id, account.description as account,
  cheque.description, date, amount from cheque
  left join party on cheque.prt_id=party.id
  left join account on cheque.acc_id=account.id
  where cheque.amount > 0 and flag is null order by date`)
	errP := DB.Select(&chequeP, `
  select cheque.rowid, prt_id, party.description as party,
  acc_id, account.description as account,
  cheque.description, date, amount from cheque
  left join party on cheque.prt_id=party.id
  left join account on cheque.acc_id=account.id
  where cheque.amount < 0 and flag is null order by date`)
	if errR == nil && errP == nil {
		c.JSON(http.StatusOK, gin.H{"R": chequeR, "P": chequeP})
	} else {
		fmt.Print(errR, errP)
		c.JSON(http.StatusBadRequest, errR)
	}
}

func chequehonor(c *gin.Context) {
	cheque := Cheques{}
	if err := c.BindJSON(&cheque); err != nil {
		c.JSON(http.StatusBadRequest, err)
		fmt.Printf("%#v\n%#v", cheque, err)
		return
	}
	dt, _ := strconv.Atoi(c.Request.URL.Query().Get("date"))
	var tType string
	usr_id := c.MustGet("usr_id").(string)

	tx, err := DB.Begin()
	if err != nil {
		goto error
	}
	defer tx.Rollback()

	if cheque.Amount > 0 {
		tType = "S"
	} else if cheque.Amount < 0 {
		tType = "P"
	}
	_, err = tx.Exec(`insert into pmttran(type, date, prt_id, acc_id,
  amount, comment, usr_id) values(?,?,?,?,?,?,?)`, tType, dt,
		cheque.Prt_id, cheque.Acc_id, cheque.Amount, cheque.Description, usr_id)
	if err != nil {
		goto error
	}
	_, err = tx.Exec(`update party set balance=balance + ?, chq_amt=chq_amt
 + ? where id=?`, cheque.Amount*-1, cheque.Amount*-1, cheque.Prt_id)
	if err != nil {
		goto error
	}
	_, err = tx.Exec(`update cheque set flag='c' where rowid=?`, cheque.Rid)
	if err != nil {
		goto error
	}
	_, err = tx.Exec(`update account set balance=balance + ? where id=?`,
		cheque.Amount, cheque.Acc_id)
	if err != nil {
		goto error
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
	}
	return
error:
	fmt.Println(err)
	c.JSON(http.StatusInternalServerError, err)
}

func chequecancel(c *gin.Context) {
	cheque := Cheques{}
	if err := c.BindJSON(&cheque); err != nil {
		c.JSON(http.StatusBadRequest, err)
		fmt.Printf("%#v\n%#v", cheque, err)
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		goto error
	}
	defer tx.Rollback()

	_, err = tx.Exec(`update party set chq_amt=chq_amt
 + ? where id=?`, cheque.Amount*-1, cheque.Prt_id)
	if err != nil {
		goto error
	}
	_, err = tx.Exec(`update cheque set flag='d!' where rowid=?`, cheque.Rid)
	if err != nil {
		goto error
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
	}
	return
error:
	fmt.Println(err)
	c.JSON(http.StatusInternalServerError, err)
}
