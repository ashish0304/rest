package main

import (
  "fmt"
  "errors"
  "database/sql"
  _"net/http"
  "github.com/gin-gonic/gin"
)

type Stock struct{
  Itm_id uint32 `json:"itm_id"`
  Quantity int32 `json:"quantity"`
  Rate float32 `json:"rate"`
  Tax float32 `json:"tax"`
  Value float32 `json:"value"`
  Cost float32 `json:"cost"`
}

type Stktran struct{
  Type string `json:"type"`
  Date string `json:"date"`
  Lcn_id uint32 `json:"lcn_id"`
  PrtAcc_id int32 `json:"prt_id"`
  Tgt_lcn_id uint32 `json:"tgt_lcn_id"`
  Expense float32 `json:"expense"`
  Prt_exp float32 `json:"prt_exp"`
  Total float32 `json:"total"`
  Flg_cost bool `json:"flg_cost"`
  Flg_merge bool `json:"flg_merge"`
  Flg_total bool `json:"flg_total"`
  Stocks []Stock `json:"stocks"`
}

type Stkrep struct {
  Date string `json:"date" db:"date"`
  Itm_id uint32 `json:"itm_id" db:"itm_id"`
  Description string `json:"description" db:"description"`
  Quantity int32 `json:"quantity" db:"quantity"`
  Rate float32 `json:"rate" db:"rate"`
}

type Amtrep struct {
  Type string `json:"type" db:"type"`
  Date string `json:"date" db:"date"`
  Prt_id NullInt64 `json:"prt_id" db:"prt_id"`
  Party NullString `json:"party" db:"party"`
  Comment NullString `json:"comment" db:"comment"`
  Amount float32 `json:"amount" db:"amount"`
}

type Locnstat struct {
  Type string `json:"type" db:"type"`
  Amount float32 `json:"amount" db:"amount"`
  Tax NullFloat64 `json:"tax" db:"tax"`
}

func stktran(c *gin.Context) {
  stktran := Stktran{}
  if err := c.BindJSON(&stktran); err != nil {
    c.AbortWithError(406, err)
    return
  }
  if stktran.Expense < 0 || stktran.Prt_exp < 0 {
    c.AbortWithError(406, errors.New("Expense, party expense must not be less than zero"))
    return
  }
  var mType = map[string]string {
    "S": "qty_sold",
    "P": "qty_recd",
    "T": "qty_tsfr",
    "A": "qty_adjt"}
  if _, ok := mType[stktran.Type]; !ok {
    c.AbortWithError(406, errors.New("Unknown transaction types"))
    return
  }
  //check for dummy locations
  var dummyLcn, dummyTgt bool
  DB.QueryRow("SELECT dummy FROM location WHERE id=?", stktran.Lcn_id).Scan(&dummyLcn)
  DB.QueryRow("SELECT dummy FROM location WHERE id=?", stktran.Tgt_lcn_id).Scan(&dummyTgt)

  //check for dummy locations and transfer
  if (dummyLcn || dummyTgt) && stktran.Type=="T" {
    c.AbortWithError(406, errors.New("transfer is not allowed to/from dummy location"))
    return
  }
  if stktran.Tgt_lcn_id < 1 && (stktran.Type == "T" || stktran.Flg_merge == true) {
    c.AbortWithError(406, errors.New("target location not specified"))
    return
  }
  if stktran.Type == "A" && stktran.Tgt_lcn_id > 0 {
    stktran.Tgt_lcn_id = 0
  }
  //get last id from location
  tid := getLastID(stktran.Lcn_id, stktran.Type) + 1
  
  //initialize transaction value
  var fTranValue float32=0

  var acc, prt uint32
  if stktran.PrtAcc_id < 0 {
    acc = uint32(stktran.PrtAcc_id * -1)
  } else if stktran.PrtAcc_id == 0 {
    acc = 1
  } else {
    prt = uint32(stktran.PrtAcc_id)
  }
  if stktran.Type=="A" || stktran.Type=="T"{
    prt = 0
    acc = 0
    stktran.Expense = 0
    stktran.Prt_exp = 0
    stktran.Flg_cost = false  //update item cost
    stktran.Flg_merge = false  //merge transaction in target location
    stktran.Flg_total = false  //update total to party/account
  }
  if !dummyLcn {
    stktran.Flg_merge = false
    stktran.Flg_total = false
  }
  if stktran.Type != "P" {
    stktran.Prt_exp = 0
    stktran.Flg_cost = false
  }
  fCostPerc := (((stktran.Expense + stktran.Prt_exp)/(stktran.Total/100))/100) + 1
  //fmt.Print(fCostPerc)
  //return
  qStktran := `INSERT INTO stktran(id, type,
               date, lcn_id, prt_id, itm_id,
               quantity, rate, tax, value, cost)
               VALUES(?,?,datetime(?),?,?,?,?,?,?,?,?)`
  qPmttran := `insert into pmttran (txn_id, type, date, acc_id,
               prt_id, amount) values(?,?,datetime(?),?,?,?)`

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
  if err != nil {goto error}
  defer tx.Rollback()
  
  stStktran, err = tx.Prepare(qStktran)
  if err != nil {goto error}
  stPmttran, err = tx.Prepare(qPmttran)
  if err != nil {goto error}
  stStockUpd, err = tx.Prepare(qStockUpd)
  if err != nil {goto error}
  stStockIns, err = tx.Prepare(qStockIns)
  if err != nil {goto error}
  stItmUpd, err = tx.Prepare(qItmUpd)
  if err != nil {goto error}
  stPrtUpd, err = tx.Prepare(qPrtUpd)
  if err != nil {goto error}
  stLcnUpd, err = tx.Prepare(qLcnUpd)
  if err != nil {goto error}
  stAccUpd, err = tx.Prepare(qAccUpd)
  if err != nil {goto error}

  for _, stk := range stktran.Stocks {
    _, err = stStktran.Exec(tid, stktran.Type, stktran.Date,
              stktran.Lcn_id, NullZero(prt), stk.Itm_id, stk.Quantity,
              stk.Rate, stk.Tax, stk.Value, stk.Cost)
    if err != nil {goto error}
    if stktran.Tgt_lcn_id > 0 {
      tQuantity := stk.Quantity
      tTax      := stk.Tax
      tValue    := stk.Value
      tCost     := stk.Cost
      if stktran.Type == "T" {
        tQuantity*=-1
        tTax     *=-1
        tValue   *=-1
        tCost    *=-1
      }
      _, err = stStktran.Exec(tid, stktran.Type, stktran.Date,
                stktran.Tgt_lcn_id, NullZero(prt), stk.Itm_id, tQuantity,
                stk.Rate, tTax, tValue, tCost)
      if err != nil {goto error}
    }
    switch stktran.Type {
    case "S":
      _, err = stStockUpd.Exec(stk.Quantity*-1, stk.Quantity, stktran.Lcn_id, stk.Itm_id)
      if err != nil {goto error}
      _, err = stStockIns.Exec(stktran.Lcn_id, stk.Itm_id, stk.Quantity*-1, stk.Quantity)
      if err != nil {goto error}
      if stktran.Flg_merge {
        _, err = stStockUpd.Exec(stk.Quantity*-1, stk.Quantity, stktran.Tgt_lcn_id, stk.Itm_id)
        if err != nil {goto error}
        _, err = stStockIns.Exec(stktran.Tgt_lcn_id, stk.Itm_id, stk.Quantity*-1, stk.Quantity)
        if err != nil {goto error}
      }
      
      fTranValue += stk.Value + stk.Tax
    case "P":
      _, err = stStockUpd.Exec(stk.Quantity, stk.Quantity, stktran.Lcn_id, stk.Itm_id)
      if err != nil {goto error}
      _, err = stStockIns.Exec(stktran.Lcn_id, stk.Itm_id, stk.Quantity, stk.Quantity)
      if err != nil {goto error}
      if stktran.Flg_merge {
        _, err = stStockUpd.Exec(stk.Quantity, stk.Quantity, stktran.Tgt_lcn_id, stk.Itm_id)
        if err != nil {goto error}
        _, err = stStockIns.Exec(stktran.Tgt_lcn_id, stk.Itm_id, stk.Quantity, stk.Quantity)
        if err != nil {goto error}
      }
      if stktran.Flg_cost {
        _, err = stItmUpd.Exec(stk.Rate * fCostPerc, stk.Itm_id)
      }
      if err != nil {goto error}

      fTranValue -= stk.Value + stk.Tax
    case "T":
      _, err = stStockUpd.Exec(stk.Quantity, stk.Quantity, stktran.Lcn_id, stk.Itm_id)
      if err != nil {goto error}
      _, err = stStockUpd.Exec(stk.Quantity*-1, stk.Quantity*-1, stktran.Tgt_lcn_id, stk.Itm_id)
      if err != nil {goto error}
      _, err = stStockIns.Exec(stktran.Lcn_id, stk.Itm_id, stk.Quantity, stk.Quantity)
      if err != nil {goto error}
      _, err = stStockIns.Exec(stktran.Tgt_lcn_id, stk.Itm_id, stk.Quantity*-1, stk.Quantity*-1)
      if err != nil {goto error}
    case "A":
      _, err = stStockUpd.Exec(stk.Quantity, stk.Quantity, stktran.Lcn_id, stk.Itm_id)
      if err != nil {goto error}
      _, err = stStockIns.Exec(stktran.Lcn_id, stk.Itm_id, stk.Quantity, stk.Quantity)
      if err != nil {goto error}
    }
  }
  //increase last transaction id in location
  _, err = stLcnUpd.Exec(stktran.Lcn_id)
  if err != nil {goto error}
  if stktran.Tgt_lcn_id > 0 {
    _, err = stLcnUpd.Exec(stktran.Tgt_lcn_id)
    if err != nil {goto error}
  }

  //update party balance

  if fTranValue != 0 {
    exp := stktran.Expense + stktran.Prt_exp
    if acc > 0 {
      _, err = stAccUpd.Exec(fTranValue + exp * -1, acc)
      if err != nil {goto error}
      _, err = stPmttran.Exec(tid, stktran.Type, stktran.Date, acc, nil, fTranValue)
      if err != nil {goto error}
      if exp != 0 {
        _, err = stPmttran.Exec(tid, "B", stktran.Date, acc, nil, exp * -1)
        if err != nil {goto error}
      }
    } else if !dummyLcn || (dummyLcn && stktran.Flg_total) {
      if stktran.Type == "P" {
        stktran.Prt_exp *= -1
      } else {
        stktran.Prt_exp = 0
      }
      _, err = stPrtUpd.Exec(fTranValue + stktran.Prt_exp, prt)
      if err != nil {goto error}
      if stktran.Expense != 0 {
        _, err = stPmttran.Exec(tid, "B", stktran.Date, 1, prt, stktran.Expense * -1)
        if err != nil {goto error}
      }
      if stktran.Prt_exp != 0 {
        _, err = stPmttran.Exec(tid, "B", stktran.Date, nil, prt, stktran.Prt_exp)
        if err != nil {goto error}
      }
    }
  }
  err = tx.Commit()
  if err != nil {
    fmt.Println(err)
    tx.Rollback()
    c.AbortWithError(500, err)
  }
  stStktran.Close(); stPmttran.Close()
  stStockUpd.Close(); stStockIns.Close()
  stItmUpd.Close(); stPrtUpd.Close()
  stLcnUpd.Close(); stAccUpd.Close()
  return
error:
  //fmt.Printf("%#v\n", err)
  c.AbortWithError(406, err)
}

func getLastID(l uint32, t string) uint32 {
  var id uint32
  DB.QueryRow(fmt.Sprintf(`select lst_%sid from location where id=?`,t), l).Scan(&id)
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
  err1 := DB.Select(&repCsSale, `select strftime('%d-%m-%Y', date) as date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='S' and
            prt_id is null and lcn_id=? and strftime('%Y-%m-%d', date) between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
  err2 := DB.Select(&repCrSale, `select strftime('%d-%m-%Y', date) as date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='S' and
            prt_id is not null and lcn_id=? and strftime('%Y-%m-%d', date) between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
  err3 := DB.Select(&repCsPurc, `select strftime('%d-%m-%Y', date) as date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='P' and
            prt_id is null and lcn_id=? and strftime('%Y-%m-%d', date) between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
  err4 := DB.Select(&repCrPurc, `select strftime('%d-%m-%Y', date) as date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='P' and
            prt_id is not null and lcn_id=? and strftime('%Y-%m-%d', date) between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
  err5 := DB.Select(&repTsfr, `select strftime('%d-%m-%Y', date) as date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='T' and
            lcn_id=? and strftime('%Y-%m-%d', date) between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
  err6 := DB.Select(&repAdjt, `select strftime('%d-%m-%Y', date) as date,
            itm_id, item.description, quantity, stktran.rate from stktran
            left join item on stktran.itm_id=item.id where type='A' and
            lcn_id=? and strftime('%Y-%m-%d', date) between ? and ?
            order by stktran.date desc`, id, dtfr, dtto)
  err7 := DB.Select(&repAmPaid, `select type, strftime('%d-%m-%Y', date) as date, prt_id,
            party.description as party, comment, amount from pmttran left join party on
            pmttran.prt_id=party.id where type in("P", "B", "W", "C", "D", "G") and
            strftime('%Y-%m-%d', date) between ? and ? order by pmttran.date desc`, dtfr, dtto)
  err8 := DB.Select(&repAmRecd, `select type, strftime('%d-%m-%Y', date) as date, prt_id,
            party.description as party, comment, amount from pmttran left join party on
            pmttran.prt_id=party.id where type in("S", "T", "H") and
            strftime('%Y-%m-%d', date) between ? and ? order by pmttran.date desc`, dtfr, dtto)
  if err1 == nil && err2 == nil && err3 == nil && err4 == nil && err5 == nil && err6 == nil && err7 == nil && err8 == nil {
    c.JSON(200, gin.H{"cssale": repCsSale,
                      "crsale": repCrSale,
                      "cspurc": repCsPurc,
                      "crpurc": repCrPurc,
                      "tsfr": repTsfr,
                      "adjt": repAdjt,
                      "ampaid": repAmPaid,
                      "amrecd": repAmRecd})
  } else {
    fmt.Print(err1, err2, err3, err4,
              err5, err6, err7, err8)
  }
}

func replcnstat(c *gin.Context) {
  repLocn := []Locnstat{}
  locn := c.Request.URL.Query().Get("locn")
  mnth := c.Request.URL.Query().Get("month")
  err := DB.Select(&repLocn, `
         select (type || ':' || round((tax/(value/100)),1)) as type,
         round(sum(value), 2) as amount, round(sum(tax), 2) as tax from stktran
         where lcn_id=? and strftime('%Y-%m', date)=?
         group by type, round((tax/(value/100)),1)
         order by type
         `, locn, mnth)
  if err == nil {
    c.JSON(200, repLocn)
  } else {
    fmt.Println(err)
  }
}
