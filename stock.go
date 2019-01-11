package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Stkmast struct {
	Lcn_id   uint `db:"lcn_id" json:"lcn_id" `
	Itm_id   uint `db:"itm_id" json:"itm_id" `
	Quantity int  `db:"quantity" json:"quantity" `
	Qty_sold int  `db:"qty_sold" json:"qty_sold" `
	Qty_recd int  `db:"qty_recd" json:"qty_recd" `
	Qty_tsfr int  `db:"qty_tsfr" json:"qty_tsfr" `
	Qty_adjt int  `db:"qty_adjt" json:"qty_adjt" `
}

type Stocks struct {
	Itm_id      uint       `db:"itm_id" json:"itm_id"`
	Description string     `db:"description" json:"description"`
	HSN         NullString `db:"hsn" json:"hsn"`
	Quantity    int        `db:"quantity" json:"quantity"`
	Cost        float32    `db:"cost" json:"cost"`
}

type Inventory struct {
	Id          NullString `db:"id" json:"id"`
	Lcn_id      NullInt64  `db:"lcn_id" json:"lcn_id" `
	Itm_id      NullInt64  `db:"itm_id" json:"itm_id" `
	Description NullString `db:"description" json:"description"`
	Quantity    NullInt64  `db:"quantity" json:"quantity" `
}

func stock(c *gin.Context) {
	stock := []Stkmast{}
	lc := c.Param("lc")
	it := c.Param("it")
	err := DB.Select(&stock, "select * from stock where lcn_id=? and itm_id=?", lc, it)
	if err == nil {
		c.JSON(http.StatusOK, stock)
	} else {
		c.JSON(http.StatusBadRequest, err)
	}
}

func stocks(c *gin.Context) {
	stock := []Stocks{}
	var sTax string
	lc := c.Param("lc")
	ta, err := strconv.ParseFloat(c.Param("ta"), 32)
	if err == nil {
		sTax = fmt.Sprintf(" and tax=%.2f", ta)
	}
	sTax += " order by description"
	err = DB.Select(&stock, `select stock.itm_id, item.description,
           item.hsn, stock.quantity, coalesce(a.rate, item.cost) as cost
           from stock left join item on stock.itm_id=item.id
           left join (select max(date),itm_id,rate from stktran
             where lcn_id=? and type='P' group by itm_id) as 'a' on stock.itm_id=a.itm_id
           where stock.quantity != 0 and lcn_id=?`+sTax, lc, lc)
	if err == nil {
		c.JSON(http.StatusOK, stock)
	} else {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
	}
}

func invs(c *gin.Context) {
	id := c.Request.URL.Query().Get("inv_id")
	lcn_id := c.Request.URL.Query().Get("lcn_id")
	var strQ string
	if len(id) == 0 {
		strQ = "select distinct id from inventory"
	} else {
		strQ = `select inventory.id, lcn_id, itm_id,
            description, quantity from inventory
            left join item on inventory.itm_id=item.id
            where inventory.id='` + id + "' and lcn_id=" + lcn_id
	}
	outs := []Inventory{}
	err := DB.Select(&outs, strQ)
	if err == nil {
		c.JSON(http.StatusOK, outs)
	} else {
		fmt.Println(err, strQ)
		c.JSON(http.StatusBadRequest, err)
	}
}

func invsadd(c *gin.Context) {
	invs := Inventory{}
	outs := []Inventory{}
	if err := c.BindJSON(&invs); err != nil {
		fmt.Println(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res, err := DB.NamedExec("update inventory set quantity="+
		"quantity+:quantity where id=:id and "+
		"lcn_id=:lcn_id and itm_id=:itm_id", &invs)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	upd, _ := res.RowsAffected()
	if upd > 0 {
		DB.Select(&outs, `select inventory.id, lcn_id, itm_id,
            description, quantity from inventory
            left join item on inventory.itm_id=item.id
            where inventory.id=? and lcn_id=?`, invs.Id.String, invs.Lcn_id.Int64)
		c.JSON(http.StatusOK, outs)
		return
	}
	_, err = DB.NamedExec("insert or ignore into inventory"+
		"(id, lcn_id, itm_id, quantity)"+
		" values(:id, :lcn_id, :itm_id, :quantity)", &invs)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
	}
}

func invsdel(c *gin.Context) {
	invs := Inventory{}
	if err := c.BindJSON(&invs); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	_, err := DB.NamedExec("delete from inventory where id=:id and "+
		"lcn_id=:lcn_id and itm_id=:itm_id", &invs)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
	}
}

func clrstk(c *gin.Context) {
	lcn_id := c.Request.URL.Query().Get("lcn_id")
	_, err := DB.Exec("delete from stock where lcn_id=?", lcn_id)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
	}
}

func expinv(c *gin.Context) {
	inv_id := c.Request.URL.Query().Get("inv_id")
	lcn_id := c.Request.URL.Query().Get("lcn_id")

	invs := []Inventory{}
	_ = DB.Select(&invs, "select itm_id, quantity from inventory where id=? and lcn_id=?", inv_id, lcn_id)

	fmt.Println(invs)
	tx, err := DB.Begin()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
	}
	defer tx.Rollback()

	stStk, err := tx.Prepare("insert into stock(lcn_id, itm_id, quantity) values(?, ?, ?)")
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
	}

	for _, s := range invs {
		_, err = stStk.Exec(lcn_id, s.Itm_id, s.Quantity)

		if err != nil {
			fmt.Println(lcn_id, s.Itm_id, err)
			c.JSON(http.StatusBadRequest, err)
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	stStk.Close()
}

func clrinv(c *gin.Context) {
	inv_id := c.Request.URL.Query().Get("inv_id")
	lcn_id := c.Request.URL.Query().Get("lcn_id")
	_, err := DB.Exec("delete from inventory where id=? and lcn_id=?", inv_id, lcn_id)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
	}
}
