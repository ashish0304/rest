package main

import (
  "fmt"
  "strconv"
  "github.com/gin-gonic/gin"
)

type Stkmast struct {
  Lcn_id uint `db:"lcn_id" json:"lcn_id" `
  Itm_id uint `db:"itm_id" json:"itm_id" `
  Quantity int `db:"quantity" json:"quantity" `
  Qty_sold int `db:"qty_sold" json:"qty_sold" `
  Qty_recd int `db:"qty_recd" json:"qty_recd" `
  Qty_tsfr int `db:"qty_tsfr" json:"qty_tsfr" `
  Qty_adjt int `db:"qty_adjt" json:"qty_adjt" `
}

type Stocks struct {
  Itm_id uint `db:"itm_id" json:"itm_id"`
  Description string `db:"description" json:"description"`
  HSN NullString `db:"hsn" json:"hsn"`
  Quantity int `db:"quantity" json:"quantity"`
  Cost float32 `db:"cost" json:"cost"`
}

func stock(c *gin.Context) {
  stock:=[]Stkmast{}
  lc := c.Param("lc")
  it := c.Param("it")
  err := DB.Select(&stock, "select * from stock where lcn_id=? and itm_id=?", lc, it)
  if err == nil{
    c.JSON(200, stock)
  }else{
    c.JSON(404, err)
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
           where stock.quantity != 0 and lcn_id=?` + sTax, lc, lc)
  if err == nil{
    c.JSON(200, stock)
  }else{
    fmt.Println(err)
    c.JSON(404, err)
  }
}
//select max(date),itm_id,rate from stktran where lcn_id=1 and type='P' group by itm_id;
