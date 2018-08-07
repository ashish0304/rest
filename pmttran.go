package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Pmttran struct {
	Type       string    `db:"type" json:"type"`
	Date       int64     `db:"date" json:"date"`
	Prt_id     uint32    `db:"prt_id" json:"prt_id"`
	Txn_id     uint32    `db:"txn_id" json:"txn_id"`
	Acc_id     uint32    `db:"acc_id" json:"acc_id"`
	Amount     float32   `db:"amount" json:"amount"`
	Chq_date   NullInt64 `db:"chq_date" json:"chq_date"`
	Comment    string    `db:"comment" json:"comment"`
	Usr_id     string    `db:"usr_id" json:"usr_id"`
	Tgt_acc_id uint32    `db:"tgt_acc_id" json:"tgt_acc_id"`
}

type Payments struct {
	Type    string     `db:"type" json:"type"`
	Date    int64      `db:"date" json:"date"`
	Prt_id  NullInt64  `db:"prt_id" json:"prt_id"`
	Party   NullString `db:"party" json:"party"`
	Account NullString `db:"account" json:"account"`
	Amount  float32    `db:"amount" json:"amount"`
	Comment NullString `db:"comment" json:"comment"`
}

type PartyPmts struct {
	Type    string     `db:"type" json:"type"`
	Date    int64      `db:"date" json:"date"`
	Account NullString `db:"account" json:"account"`
	Amount  float32    `db:"amount" json:"amount"`
	Comment NullString `db:"comment" json:"comment"`
}

type Acctrans struct {
	Type    string     `db:"type" json:"type"`
	Date    int64      `db:"date" json:"date"`
	Prt_id  NullInt64  `db:"prt_id" json:"prt_id"`
	Party   NullString `db:"party" json:"party"`
	Amount  float32    `db:"amount" json:"amount"`
	Comment NullString `db:"comment" json:"comment"`
}

type Ptran struct {
	Rid     int64      `db:"rowid" json:"rowid"`
	Type    string     `db:"type" json:"type"`
	Date    int64      `db:"date" json:"date"`
	Prt_id  NullInt64  `db:"prt_id" json:"prt_id"`
	Txn_id  NullInt64  `db:"txn_id" json:"txn_id"`
	Acc_id  NullInt64  `db:"acc_id" json:"acc_id"`
	Amount  float32    `db:"amount" json:"amount"`
	Comment NullString `db:"comment" json:"comment"`
	Usr_id  NullString `db:"usr_id" json:"usr_id"`
	Flag    NullString `db:"flag" json:"flag"`
}

func pmttran(c *gin.Context) {
	pmttran := Pmttran{}
	if err := c.BindJSON(&pmttran); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		fmt.Printf("%#v \n%#v", pmttran, err)
		return
	}
	if pmttran.Amount <= 0 || pmttran.Acc_id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount/Account must not be 0 or less!"})
		return
	}
	if (pmttran.Type == "S" || pmttran.Type == "P" ||
		pmttran.Type == "G" || pmttran.Type == "H") && pmttran.Prt_id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Party is required for submitted transaction!"})
		return
	}
	if pmttran.Type != "T" {
		pmttran.Tgt_acc_id = 0
	}
	if pmttran.Type == "T" || pmttran.Type == "W" ||
		pmttran.Type == "C" || pmttran.Type == "D" {
		pmttran.Prt_id = 0
	}
	qPmttran := `insert into pmttran(type, date, prt_id, acc_id,
               amount, comment, usr_id) values(?,?,?,?,?,?,?)`
	qPrtUpd := `update party set balance=balance + ? where id=?`
	qChqUpd := `update party set chq_amt=chq_amt + ? where id=?`
	qChqCrt := `insert into cheque(acc_id, prt_id, description, date, amount)
             values(?, ?, ?, ?, ?)`
	qAccUpd := `update account set balance=balance + ? where id=?`
	var stPmttran, stPrtUpd, stAccUpd, stChqUpd, stChqCrt *sql.Stmt

	//get usr id from context
	pmttran.Usr_id = c.MustGet("usr_id").(string)
 fmt.Println(pmttran)
	tx, err := DB.Begin()
	if err != nil {
		goto error
	}
	defer tx.Rollback()

	stPmttran, err = tx.Prepare(qPmttran)
	if err != nil {
		goto error
	}
	stPrtUpd, err = tx.Prepare(qPrtUpd)
	if err != nil {
		goto error
	}
	stChqUpd, err = tx.Prepare(qChqUpd)
	if err != nil {
		goto error
	}
	stChqCrt, err = tx.Prepare(qChqCrt)
	if err != nil {
		goto error
	}
	stAccUpd, err = tx.Prepare(qAccUpd)
	if err != nil {
		goto error
	}

	if pmttran.Chq_date.Int64 == 0 {
		switch pmttran.Type {
		case "S", "T", "H": //Sale, Transfer/Deposit, Petty Loan Taken
			_, err = stPmttran.Exec(pmttran.Type, pmttran.Date,
				NullZero(pmttran.Prt_id), pmttran.Acc_id,
				pmttran.Amount, pmttran.Comment, pmttran.Usr_id)
			if err != nil {
				goto error
			}
			_, err = stAccUpd.Exec(pmttran.Amount, pmttran.Acc_id)
			if err != nil {
				goto error
			}

			if pmttran.Tgt_acc_id > 0 {
				_, err = stPmttran.Exec(pmttran.Type, pmttran.Date,
					NullZero(pmttran.Prt_id), pmttran.Tgt_acc_id,
					pmttran.Amount*-1, pmttran.Comment, pmttran.Usr_id)
				if err != nil {
					goto error
				}
				_, err = stAccUpd.Exec(pmttran.Amount*-1, pmttran.Tgt_acc_id)
				if err != nil {
					goto error
				}
			}
		case "P", "B", "W", "C", "D", "G": //Purchase, Bus/Transport, Wages/Rent, Service Charges, Petty Expenses, Petty Loan Given
			_, err = stPmttran.Exec(pmttran.Type, pmttran.Date,
				NullZero(pmttran.Prt_id), pmttran.Acc_id,
				pmttran.Amount*-1, pmttran.Comment, pmttran.Usr_id)
			if err != nil {
				goto error
			}
			_, err = stAccUpd.Exec(pmttran.Amount*-1, pmttran.Acc_id)
			if err != nil {
				goto error
			}
		}
	}

	switch pmttran.Type {
	case "S", "H": //Sale, Petty Loan Taken
		if pmttran.Chq_date.Int64 != 0 {
			_, err = stChqUpd.Exec(pmttran.Amount, pmttran.Prt_id)
			if err != nil {
				goto error
			}
			_, err = stChqCrt.Exec(pmttran.Acc_id, pmttran.Prt_id, pmttran.Comment, pmttran.Chq_date, pmttran.Amount)
			if err != nil {
				goto error
			}
		} else {
			_, err = stPrtUpd.Exec(pmttran.Amount*-1, pmttran.Prt_id)
			if err != nil {
				goto error
			}
		}
	case "P", "B", "G": //Purchase, Bus/Transport, Petty Loan Given
		if pmttran.Chq_date.Int64 != 0 {
			_, err = stChqUpd.Exec(pmttran.Amount*-1, pmttran.Prt_id)
			if err != nil {
				goto error
			}
			_, err = stChqCrt.Exec(pmttran.Acc_id, pmttran.Prt_id, pmttran.Comment, pmttran.Chq_date, pmttran.Amount*-1)
			if err != nil {
				goto error
			}
		} else {
			_, err = stPrtUpd.Exec(pmttran.Amount, pmttran.Prt_id)
			if err != nil {
				goto error
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
	}
	stPmttran.Close()
	stPrtUpd.Close()
	stChqUpd.Close()
	stChqCrt.Close()
	stAccUpd.Close()
	return
error:
	fmt.Println(err)
	c.JSON(500, gin.H{"error": err})
}

func payments(c *gin.Context) {
	pmts := []Payments{}
	offset, e1 := strconv.Atoi(c.Request.URL.Query().Get("offset"))
	if e1 != nil {
		offset = -1
	}
	limit, e2 := strconv.Atoi(c.Request.URL.Query().Get("limit"))
	if e2 != nil {
		limit = -1
	}
	err := DB.Select(&pmts, `select type, date, prt_id, party.description as party, 
         account.description as account, amount, comment
         from pmttran left join account on acc_id=account.id
         left join party on prt_id=party.id order by date desc limit ?,?`, offset, limit)
	if err != nil {
		c.JSON(400, err)
		fmt.Println(err)
	} else {
		c.JSON(200, pmts)
	}
}

func prtpayments(c *gin.Context) {
	id := c.Param("id")
	offset, e1 := strconv.Atoi(c.Request.URL.Query().Get("offset"))
	if e1 != nil {
		offset = -1
	}
	limit, e2 := strconv.Atoi(c.Request.URL.Query().Get("limit"))
	if e2 != nil {
		limit = -1
	}
	pmts := []PartyPmts{}
	err := DB.Select(&pmts, `select type, date,
         account.description as account, amount, comment
         from pmttran left join account on acc_id=account.id
         where prt_id=? order by date desc limit ?,?`, id, offset, limit)
	if err != nil {
		c.JSON(400, err)
		//fmt.Println(err)
	} else {
		c.JSON(200, pmts)
	}
}

func acctrans(c *gin.Context) {
	pmts := []Acctrans{}
	acc, e0 := strconv.Atoi(c.Request.URL.Query().Get("acc"))
	if e0 != nil {
		acc = 0
	}
	offset, e1 := strconv.Atoi(c.Request.URL.Query().Get("offset"))
	if e1 != nil {
		offset = -1
	}
	limit, e2 := strconv.Atoi(c.Request.URL.Query().Get("limit"))
	if e2 != nil {
		limit = -1
	}
	DB.Select(&pmts, `select type, date,
         prt_id, party.description as party, amount, comment
         from pmttran left join party on prt_id=party.id where acc_id=?
         order by date desc limit ?,?`, acc, offset, limit)
	c.JSON(200, pmts)
}

func gptran(c *gin.Context) {
	var length int
	ptran := []Ptran{}
	ln, _ := strconv.ParseBool(c.Request.URL.Query().Get("length"))
	acc := c.Request.URL.Query().Get("acc")
	typ := c.Request.URL.Query().Get("type")
	off := c.Request.URL.Query().Get("offset")
	lim := c.Request.URL.Query().Get("limit")

	if ln {
		rw := DB.QueryRow("select count() from pmttran "+
			"where (acc_id=? or ?='') "+
			"and (type=? or ?='') "+
			"order by date desc", acc, acc, typ, typ)
		rw.Scan(&length)
	}
	err := DB.Select(&ptran, "select rowid, * from pmttran "+
		"where (acc_id=? or ?='') and (type=? or ?='') "+
		"order by date desc limit ?,?", acc, acc, typ, typ, off, lim)
	if err == nil {
		//fmt.Println(ptran)
		c.JSON(200, gin.H{"len": length, "rows": ptran})
	} else {
		fmt.Println(err)
	}
}

func pptran(c *gin.Context) {
	ptran := Ptran{}
	if err := c.BindJSON(&ptran); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		fmt.Printf("%#v \n%#v", ptran, err)
		return
	}
	_, err := DB.NamedExec("update pmttran set type=:type, date=:date, "+
		"prt_id=case when :prt_id=0 then null else :prt_id end, "+
		"txn_id=case when :txn_id=0 then null else :txn_id end, "+
		"acc_id=case when :acc_id=0 then null else :acc_id end, "+
		"usr_id=case when :usr_id='' then null else :usr_id end, "+
		"amount=:amount, comment=:comment, flag=:flag "+
		"where rowid=:rowid", &ptran)
	if err != nil {
		fmt.Println(ptran, err)
		c.JSON(400, err)
	} else {
		c.JSON(200, ptran)
	}
}
