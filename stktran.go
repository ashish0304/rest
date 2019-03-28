package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Stock struct {
	Itm_id   uint32  `json:"itm_id"`
	Quantity int32   `json:"quantity"`
	Rate     float32 `json:"rate"`
	Tax      float32 `json:"tax"`
	Value    float32 `json:"value"`
	Cost     float32 `json:"cost"`
}

type Stktran struct {
	Type       string  `json:"type"`
	Date       int64   `json:"date"`
	Lcn_id     uint32  `json:"lcn_id"`
	PrtAcc_id  int32   `json:"prt_id"`
	Tgt_lcn_id uint32  `json:"tgt_lcn_id"`
	Expense    float32 `json:"expense"`
	Prt_exp    float32 `json:"prt_exp"`
	Total      float32 `json:"total"`
	Flg_cost   bool    `json:"flg_cost"`
	Flg_merge  bool    `json:"flg_merge"`
	Flg_total  bool    `json:"flg_total"`
	Usr_id     string  `json:"usr_id"`
	Stocks     []Stock `json:"stocks"`
}

func stktran(c *gin.Context) {
	stktran := Stktran{}
	if err := c.BindJSON(&stktran); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if stktran.Expense < 0 || stktran.Prt_exp < 0 {
		c.JSON(http.StatusBadRequest, "Error: Expense, party expense must not be less than zero")
		return
	}
	var mType = map[string]string{
		"S": "qty_sold",
		"P": "qty_recd",
		"T": "qty_tsfr",
		"A": "qty_adjt"}
	if _, ok := mType[stktran.Type]; !ok {
		c.JSON(http.StatusBadRequest, "Error: Unknown transaction types")
		return
	}

	logSql.Println("==========STKTRAN==========")
	//check for dummy locations
	var dummyLcn, dummyTgt bool
	DB.QueryRow("SELECT dummy FROM location WHERE id=?", stktran.Lcn_id).Scan(&dummyLcn)
	DB.QueryRow("SELECT dummy FROM location WHERE id=?", stktran.Tgt_lcn_id).Scan(&dummyTgt)

	//check for dummy locations and transfer
	if (dummyLcn || dummyTgt) && stktran.Type == "T" {
		c.JSON(http.StatusBadRequest, "Error: Transfer is not allowed to/from dummy location")
		return
	}
	if stktran.Tgt_lcn_id < 1 && (stktran.Type == "T" || stktran.Flg_merge == true) {
		c.JSON(http.StatusBadRequest, "Error: Target location not specified")
		return
	}

	for i, _ := range stktran.Stocks {
		if stktran.Type == "T" || stktran.Type == "A" {
			stktran.Stocks[i].Rate = 0
			stktran.Stocks[i].Tax = 0
			stktran.Stocks[i].Cost = 0
			stktran.Stocks[i].Value = 0
		} else if stktran.Stocks[i].Rate == 0 || stktran.Stocks[i].Cost == 0 {
			c.JSON(http.StatusBadRequest, "Error: Rate or cost not provided for the item")
			return
		}
	}

	if stktran.Type == "A" && stktran.Tgt_lcn_id > 0 {
		stktran.Tgt_lcn_id = 0
	}
	//get last id from location
	tid := getLastID(stktran.Lcn_id, stktran.Type) + 1
	ttid := getLastID(stktran.Tgt_lcn_id, stktran.Type) + 1

	//initialize transaction value
	var fTranValue float32 = 0

	//get usr id from context
	stktran.Usr_id = c.MustGet("usr_id").(string)
	//fmt.Println(stktran.Usr_id)
	var acc, prt uint32
	if stktran.PrtAcc_id < 0 {
		acc = uint32(stktran.PrtAcc_id * -1)
		stktran.Prt_exp = 0
	} else if stktran.PrtAcc_id == 0 {
		acc = 1
		stktran.Prt_exp = 0
	} else {
		prt = uint32(stktran.PrtAcc_id)
	}
	if stktran.Type == "A" || stktran.Type == "T" {
		prt = 0
		acc = 0
		stktran.Expense = 0
		stktran.Prt_exp = 0
		stktran.Flg_cost = false  //update item cost
		stktran.Flg_merge = false //merge transaction in target location
		stktran.Flg_total = false //update total to party/account
	}
	if !dummyLcn {
		stktran.Flg_merge = false
		stktran.Flg_total = true
		if stktran.Type == "P" {
			stktran.Flg_cost = true
		}
	}
	if stktran.Type != "P" {
		stktran.Flg_cost = false
	}

	var fCostPerc float32 = 1
	if (stktran.Expense + stktran.Prt_exp) > 0 {
		fCostPerc += (((stktran.Expense + stktran.Prt_exp) / (stktran.Total / 100)) / 100)
		fCostPerc += .002 //amount transfer charges
		//fmt.Print(fCostPerc)
	}
	//return
	qStktran := `INSERT INTO stktran(id, type,
               date, lcn_id, prt_id, itm_id,
               quantity, rate, tax, value, cost, usr_id)
               VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`
	qPmttran := `insert into pmttran (txn_id, type, date, acc_id,
               prt_id, amount, usr_id) values(?,?,?,?,?,?,?)`

	qStockUpd := `update stock set quantity=quantity+?, ` + mType[stktran.Type]
	qStockUpd += `=` + mType[stktran.Type] + `+? where lcn_id=? and itm_id=?`

	qStockIns := `insert or ignore into stock(lcn_id, itm_id, quantity, `
	qStockIns += mType[stktran.Type] + `) values(?, ?, ?, ?)`

	qItmUpd := `update item set cost=? where id=?`
	qPrtUpd := `update party set balance=balance + ? where id=?`
	qLcnUpd := `update location set lst_` + stktran.Type +
		`id=lst_` + stktran.Type + `id+1 where id=?`
	qAccUpd := `update account set balance=balance + ? where id=?`

	var stStktran, stPmttran, stStockUpd,
		stStockIns, stItmUpd, stPrtUpd,
		stLcnUpd, stAccUpd *sql.Stmt

	tx, err := DB.Begin()
	if err != nil {
		goto error
	}
	defer tx.Rollback()

	stStktran, err = tx.Prepare(qStktran)
	if err != nil {
		goto error
	}
	stPmttran, err = tx.Prepare(qPmttran)
	if err != nil {
		goto error
	}
	stStockUpd, err = tx.Prepare(qStockUpd)
	if err != nil {
		goto error
	}
	stStockIns, err = tx.Prepare(qStockIns)
	if err != nil {
		goto error
	}
	stItmUpd, err = tx.Prepare(qItmUpd)
	if err != nil {
		goto error
	}
	stPrtUpd, err = tx.Prepare(qPrtUpd)
	if err != nil {
		goto error
	}
	stLcnUpd, err = tx.Prepare(qLcnUpd)
	if err != nil {
		goto error
	}
	stAccUpd, err = tx.Prepare(qAccUpd)
	if err != nil {
		goto error
	}

	for _, stk := range stktran.Stocks {
		_, err = stStktran.Exec(tid, stktran.Type, stktran.Date,
			stktran.Lcn_id, NullZero(prt), stk.Itm_id, stk.Quantity,
			stk.Rate, stk.Tax, stk.Value, stk.Cost, stktran.Usr_id)
		if err != nil {
			goto error
		}
		if stktran.Tgt_lcn_id > 0 {
			tQuantity := stk.Quantity
			tTax := stk.Tax
			tValue := stk.Value
			tCost := stk.Cost
			if stktran.Type == "T" {
				tQuantity *= -1
				tTax *= -1
				tValue *= -1
				tCost *= -1
			}
			_, err = stStktran.Exec(ttid, stktran.Type, stktran.Date,
				stktran.Tgt_lcn_id, NullZero(prt), stk.Itm_id, tQuantity,
				stk.Rate, tTax, tValue, tCost, stktran.Usr_id)
			if err != nil {
				goto error
			}
		}
		switch stktran.Type {
		case "S":
			_, err = stStockUpd.Exec(stk.Quantity*-1, stk.Quantity, stktran.Lcn_id, stk.Itm_id)
			if err != nil {
				goto error
			}
			_, err = stStockIns.Exec(stktran.Lcn_id, stk.Itm_id, stk.Quantity*-1, stk.Quantity)
			if err != nil {
				goto error
			}
			if stktran.Flg_merge {
				_, err = stStockUpd.Exec(stk.Quantity*-1, stk.Quantity, stktran.Tgt_lcn_id, stk.Itm_id)
				if err != nil {
					goto error
				}
				_, err = stStockIns.Exec(stktran.Tgt_lcn_id, stk.Itm_id, stk.Quantity*-1, stk.Quantity)
				if err != nil {
					goto error
				}
			}

			fTranValue += stk.Value + stk.Tax
		case "P":
			_, err = stStockUpd.Exec(stk.Quantity, stk.Quantity, stktran.Lcn_id, stk.Itm_id)
			if err != nil {
				goto error
			}
			_, err = stStockIns.Exec(stktran.Lcn_id, stk.Itm_id, stk.Quantity, stk.Quantity)
			if err != nil {
				goto error
			}
			if stktran.Flg_merge {
				_, err = stStockUpd.Exec(stk.Quantity, stk.Quantity, stktran.Tgt_lcn_id, stk.Itm_id)
				if err != nil {
					goto error
				}
				_, err = stStockIns.Exec(stktran.Tgt_lcn_id, stk.Itm_id, stk.Quantity, stk.Quantity)
				if err != nil {
					goto error
				}
			}
			if stktran.Flg_cost && stk.Quantity > 0 {
				_, err = stItmUpd.Exec(stk.Rate*fCostPerc, stk.Itm_id)
			}
			if err != nil {
				goto error
			}

			fTranValue -= stk.Value + stk.Tax
			stktran.Prt_exp *= -1
		case "T":
			_, err = stStockUpd.Exec(stk.Quantity, stk.Quantity, stktran.Lcn_id, stk.Itm_id)
			if err != nil {
				goto error
			}
			_, err = stStockUpd.Exec(stk.Quantity*-1, stk.Quantity*-1, stktran.Tgt_lcn_id, stk.Itm_id)
			if err != nil {
				goto error
			}
			_, err = stStockIns.Exec(stktran.Lcn_id, stk.Itm_id, stk.Quantity, stk.Quantity)
			if err != nil {
				goto error
			}
			_, err = stStockIns.Exec(stktran.Tgt_lcn_id, stk.Itm_id, stk.Quantity*-1, stk.Quantity*-1)
			if err != nil {
				goto error
			}
		case "A":
			_, err = stStockUpd.Exec(stk.Quantity, stk.Quantity, stktran.Lcn_id, stk.Itm_id)
			if err != nil {
				goto error
			}
			_, err = stStockIns.Exec(stktran.Lcn_id, stk.Itm_id, stk.Quantity, stk.Quantity)
			if err != nil {
				goto error
			}
		}
	}
	//increase last transaction id in location
	_, err = stLcnUpd.Exec(stktran.Lcn_id)
	if err != nil {
		goto error
	}
	if stktran.Tgt_lcn_id > 0 {
		_, err = stLcnUpd.Exec(stktran.Tgt_lcn_id)
		if err != nil {
			goto error
		}
	}

	//update party balance
	if fTranValue != 0 {
		stktran.Expense *= -1
		if acc > 0 {
			if stktran.Flg_total {
				_, err = stAccUpd.Exec(fTranValue+stktran.Expense, acc)
				if err != nil {
					goto error
				}
			}
			if !dummyLcn {
				_, err = stPmttran.Exec(tid, stktran.Type, stktran.Date, acc, nil, fTranValue, stktran.Usr_id)
				if err != nil {
					goto error
				}
			}
			if stktran.Expense != 0 {
				_, err = stPmttran.Exec(tid, "B", stktran.Date, acc, nil, stktran.Expense, stktran.Usr_id)
				if err != nil {
					goto error
				}
			}
		} else if !dummyLcn || (dummyLcn && stktran.Flg_total) {
			_, err = stPrtUpd.Exec(fTranValue+stktran.Prt_exp, prt)
			if err != nil {
				goto error
			}
			if stktran.Expense != 0 {
				_, err = stPmttran.Exec(tid, "B", stktran.Date, 1, nil, stktran.Expense, stktran.Usr_id)
				if err != nil {
					goto error
				}
				_, err = stAccUpd.Exec(stktran.Expense, 1)
				if err != nil {
					goto error
				}
			}
			if stktran.Prt_exp != 0 {
				_, err = stPmttran.Exec(tid, "B", stktran.Date, nil, prt, stktran.Prt_exp, stktran.Usr_id)
				if err != nil {
					goto error
				}
			}
		}
	}
	logSql.Println("**********STKTRAN**********")

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, err)
	}

	stStktran.Close()
	stPmttran.Close()
	stStockUpd.Close()
	stStockIns.Close()
	stItmUpd.Close()
	stPrtUpd.Close()
	stLcnUpd.Close()
	stAccUpd.Close()
	return
error:
	fmt.Printf("%#v\n", err)
	c.JSON(http.StatusBadRequest, err)
}

func getLastID(l uint32, t string) uint32 {
	var id uint32
	DB.QueryRow(fmt.Sprintf(`select lst_%sid from location where id=?`, t), l).Scan(&id)
	return id
}
