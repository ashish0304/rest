package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
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

func pmttran(c *gin.Context) {
	pmttran := Pmttran{}
	if err := c.BindJSON(&pmttran); err != nil {
		c.JSON(http.StatusBadRequest, err)
		fmt.Printf("%#v \n%#v", pmttran, err)
		return
	}
	if pmttran.Amount <= 0 || pmttran.Acc_id <= 0 {
		c.JSON(http.StatusBadRequest, "Error: Amount/Account must not be 0 or less!")
		return
	}
	if (pmttran.Type == "S" || pmttran.Type == "P" ||
		pmttran.Type == "G" || pmttran.Type == "H") && pmttran.Prt_id <= 0 {
		c.JSON(http.StatusBadRequest, "Error: Party is required for submitted transaction!")
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

	logSql.Println("==========PMTTRAN==========")
	//get usr id from context
	pmttran.Usr_id = c.MustGet("usr_id").(string)
	//fmt.Println(pmttran)
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
	logSql.Println("**********PMTTRAN**********")

	stPmttran.Close()
	stPrtUpd.Close()
	stChqUpd.Close()
	stChqCrt.Close()
	stAccUpd.Close()
	return
error:
	fmt.Println(err)
	c.JSON(http.StatusInternalServerError, err)
}
