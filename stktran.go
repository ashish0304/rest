package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"strconv"
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

type ItemTran struct {
	Type     string  `db:"type" json:"type"`
	Date     int64   `db:"date" json:"date"`
	Quantity int32   `db:"quantity" json:"quantity"`
	Rate     float32 `db:"rate" json:"rate"`
}

type Stkrep struct {
	Date        int64       `json:"date" db:"date"`
	Itm_id      uint32      `json:"itm_id" db:"itm_id"`
	Description string      `json:"description" db:"description"`
	Quantity    int32       `json:"quantity" db:"quantity"`
	Rate        float32     `json:"rate" db:"rate"`
	Cost        NullFloat64 `json:"cost" db:"cost"`
}

type Amtrep struct {
	Type    string     `json:"type" db:"type"`
	Date    int64      `json:"date" db:"date"`
	Prt_id  NullInt64  `json:"prt_id" db:"prt_id"`
	Party   NullString `json:"party" db:"party"`
	Comment NullString `json:"comment" db:"comment"`
	Amount  float32    `json:"amount" db:"amount"`
}

type Locnstat struct {
	Type   string      `json:"type" db:"type"`
	Amount float32     `json:"amount" db:"amount"`
	Tax    NullFloat64 `json:"tax" db:"tax"`
}

type PartyItms struct {
	Type     string  `db:"type" json:"type"`
	Date     int64   `db:"date" json:"date"`
	Item     string  `db:"item" json:"item"`
	Quantity int32   `db:"quantity" json:"quantity"`
	Rate     float32 `db:"rate" json:"rate"`
}

type Stran struct {
	Rid      uint32      `json:"rowid" db:"rowid"`
	Id       uint32      `json:"id" db:"id"`
	Type     string      `json:"type" db:"type"`
	Date     int64       `json:"date" db:"date"`
	Lcn_id   uint32      `json:"lcn_id" db:"lcn_id"`
	Prt_id   NullInt64   `json:"prt_id" db:"prt_id"`
	Itm_id   uint32      `json:"itm_id" db:"itm_id"`
	Quantity int32       `json:"quantity" db:"quantity"`
	Rate     float32     `json:"rate" db:"rate"`
	Value    float32     `json:"value" db:"value"`
	Tax      NullFloat64 `json:"tax" db:"tax"`
	Cost     NullFloat64 `json:"cost" db:"cost"`
	Usr_id   NullString  `json:"usr_id" db:"usr_id"`
	Flag     NullString  `json:"flag" db:"flag"`
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
	} else if stktran.PrtAcc_id == 0 {
		acc = 1
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
		stktran.Flg_total = false
		if stktran.Type == "P" {
			stktran.Flg_cost = true
		}
	}
	if stktran.Type != "P" {
		stktran.Prt_exp = 0
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
		exp := stktran.Expense + stktran.Prt_exp
		if acc > 0 {
			_, err = stAccUpd.Exec(fTranValue+exp*-1, acc)
			if err != nil {
				goto error
			}
			if !dummyLcn {
				_, err = stPmttran.Exec(tid, stktran.Type, stktran.Date, acc, nil, fTranValue, stktran.Usr_id)
				if err != nil {
					goto error
				}
			}
			if exp != 0 {
				_, err = stPmttran.Exec(tid, "B", stktran.Date, acc, nil, exp*-1, stktran.Usr_id)
				if err != nil {
					goto error
				}
			}
		} else if !dummyLcn || (dummyLcn && stktran.Flg_total) {
			if stktran.Type == "P" {
				_, err = stPrtUpd.Exec(fTranValue+stktran.Prt_exp*-1, prt)
			} else {
				_, err = stPrtUpd.Exec(fTranValue+stktran.Expense, prt)
			}

			if err != nil {
				goto error
			}
			if stktran.Expense != 0 {
				_, err = stPmttran.Exec(tid, "B", stktran.Date, 1, prt, stktran.Expense*-1, stktran.Usr_id)
				if err != nil {
					goto error
				}
				_, err = stAccUpd.Exec(stktran.Expense*-1, 1)
				if err != nil {
					goto error
				}
			}
			if stktran.Prt_exp != 0 {
				_, err = stPmttran.Exec(tid, "B", stktran.Date, nil, prt, stktran.Prt_exp*-1, stktran.Usr_id)
				if err != nil {
					goto error
				}
			}
		}
	}
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

func repdateitm(c *gin.Context) {
	repCsSale := []Stkrep{}
	repCrSale := []Stkrep{}
	repCsPurc := []Stkrep{}
	repCrPurc := []Stkrep{}
	repTsfr := []Stkrep{}
	repAdjt := []Stkrep{}
	repAmPaid := []Amtrep{}
	repAmRecd := []Amtrep{}
	id := c.Request.URL.Query().Get("id")
	dtfr := c.Request.URL.Query().Get("dtfr")
	dtto := c.Request.URL.Query().Get("dtto")
	err1 := DB.Select(&repCsSale, `select date,
            itm_id, item.description, quantity, stktran.rate, stktran.cost from stktran
            left join item on stktran.itm_id=item.id where type='S' and
            prt_id is null and lcn_id=? and date between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
	err2 := DB.Select(&repCrSale, `select date,
            itm_id, item.description, quantity, stktran.rate, stktran.cost from stktran
            left join item on stktran.itm_id=item.id where type='S' and
            prt_id is not null and lcn_id=? and date between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
	err3 := DB.Select(&repCsPurc, `select date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='P' and
            prt_id is null and lcn_id=? and date between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
	err4 := DB.Select(&repCrPurc, `select date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='P' and
            prt_id is not null and lcn_id=? and date between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
	err5 := DB.Select(&repTsfr, `select date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='T' and
            lcn_id=? and date between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
	err6 := DB.Select(&repAdjt, `select date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='A' and
            lcn_id=? and date between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
	err7 := DB.Select(&repAmPaid, `select type, date, prt_id,
            party.description as party, comment, amount from pmttran left join party on
            pmttran.prt_id=party.id where type in("P", "B", "W", "C", "D", "G") and
            date between ? and ? order by pmttran.date desc`, dtfr, dtto)
	err8 := DB.Select(&repAmRecd, `select type, date, prt_id,
            party.description as party, comment, amount from pmttran left join party on
            pmttran.prt_id=party.id where type in("S", "T", "H") and
            date between ? and ? order by pmttran.date desc`, dtfr, dtto)
	if err1 == nil && err2 == nil && err3 == nil && err4 == nil && err5 == nil && err6 == nil && err7 == nil && err8 == nil {
		c.JSON(http.StatusOK, gin.H{"cssale": repCsSale,
			"crsale": repCrSale,
			"cspurc": repCsPurc,
			"crpurc": repCrPurc,
			"tsfr":   repTsfr,
			"adjt":   repAdjt,
			"ampaid": repAmPaid,
			"amrecd": repAmRecd})
	} else {
		fmt.Print(err1, err2, err3, err4,
			err5, err6, err7, err8)
	}
}

func itemtran(c *gin.Context) {
	id := c.Param("id")
	offset, e1 := strconv.Atoi(c.Request.URL.Query().Get("offset"))
	if e1 != nil {
		offset = -1
	}
	limit, e2 := strconv.Atoi(c.Request.URL.Query().Get("limit"))
	if e2 != nil {
		limit = -1
	}
	trns := []ItemTran{}
	err := DB.Select(&trns, `select type, date,
         quantity, rate from stktran where itm_id=?
         order by date desc limit ?,?`, id, offset, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		//fmt.Println(err)
	} else {
		c.JSON(http.StatusOK, trns)
	}
}

func prtitems(c *gin.Context) {
	id := c.Param("id")
	offset, e1 := strconv.Atoi(c.Request.URL.Query().Get("offset"))
	if e1 != nil {
		offset = -1
	}
	limit, e2 := strconv.Atoi(c.Request.URL.Query().Get("limit"))
	if e2 != nil {
		limit = -1
	}
	itms := []PartyItms{}
	err := DB.Select(&itms, `select type, date,
         item.description as item, quantity, rate from stktran 
         left join item on itm_id=item.id
         left join location on lcn_id=location.id where dummy=0 and prt_id=?
         order by date desc limit ?,?`, id, offset, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		//fmt.Println(err)
	} else {
		c.JSON(http.StatusOK, itms)
	}
}

func replcnstat(c *gin.Context) {
	repLocn := []Locnstat{}
	locn := c.Request.URL.Query().Get("locn")
	mnth := c.Request.URL.Query().Get("month")
	err := DB.Select(&repLocn, `
         select (type || ':' || round((tax/(value/100)),1)) as type,
         round(sum(value), 2) as amount, round(sum(tax), 2) as tax from stktran
         where lcn_id=? and strftime('%Y-%m', date, 'unixepoch')=? and value != 0
         group by type, round((tax/(value/100)),1) order by type`, locn, mnth)
	if err == nil {
		c.JSON(http.StatusOK, repLocn)
	} else {
		fmt.Println(err)
	}
}

func gstran(c *gin.Context) {
	var length int
	stran := []Stran{}
	ln, _ := strconv.ParseBool(c.Request.URL.Query().Get("length"))
	id := c.Request.URL.Query().Get("id")
	lcn := c.Request.URL.Query().Get("lcn")
	typ := c.Request.URL.Query().Get("type")
	off := c.Request.URL.Query().Get("offset")
	lim := c.Request.URL.Query().Get("limit")

	if ln {
		rw := DB.QueryRow("select count() from stktran "+
			"where (id=? or ?='') "+
			"and (type=? or ?='') "+
			"and (lcn_id=? or ?='') order by date desc", id, id, typ, typ, lcn, lcn)
		rw.Scan(&length)
	}
	err := DB.Select(&stran, "select rowid, * from stktran "+
		"where (id=? or ?='') and (type=? or ?='') "+
		"and (lcn_id=? or ?='') order by date desc limit ?,?", id, id, typ, typ, lcn, lcn, off, lim)
	if err == nil {
		//fmt.Println(stran)
		c.JSON(http.StatusOK, gin.H{"len": length, "rows": stran})
	} else {
		fmt.Println(err)
	}
}

func pstran(c *gin.Context) {
	stran := Stran{}
	if err := c.BindJSON(&stran); err != nil {
		c.JSON(http.StatusBadRequest, err)
		fmt.Printf("%#v \n%#v", stran, err)
		return
	}
	_, err := DB.NamedExec("update stktran set id=:id, "+
		"type=:type, date=:date, lcn_id=:lcn_id, "+
		"prt_id=case when :prt_id=0 then null else :prt_id end, "+
		"usr_id=case when :usr_id='' then null else :usr_id end, "+
		"itm_id=:itm_id, quantity=:quantity, "+
		"rate=:rate, value=:value, tax=:tax, "+
		"cost=:cost, flag=:flag where rowid=:rowid", &stran)
	if err != nil {
		fmt.Println(stran, err)
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, stran)
	}
}
