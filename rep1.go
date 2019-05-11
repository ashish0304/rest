package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

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
            party.description as party, acc_id,
            account.description as account,
            comment, amount from pmttran
            left join party on pmttran.prt_id=party.id
            left join account on pmttran.acc_id=account.id
            where type in("P", "B", "W", "C", "D", "G") and
            date between ? and ? order by pmttran.date desc`, dtfr, dtto)
	err8 := DB.Select(&repAmRecd, `select type, date, prt_id,
            party.description as party,
            account.description as account,
            comment, amount from pmttran
            left join party on pmttran.prt_id=party.id
            left join account on pmttran.acc_id=account.id
            where type in("S", "T", "H") and
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
	trns := []ItemTran{}
	err := DB.Select(&trns, `select type, date,
         quantity, rate from stktran where itm_id=?
         order by date desc`, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		//fmt.Println(err)
	} else {
		c.JSON(http.StatusOK, trns)
	}
}

func prtitems(c *gin.Context) {
	id := c.Param("id")
	lc := c.Request.URL.Query().Get("locn")
	tp := c.Request.URL.Query().Get("type")
	date, _ := strconv.Atoi(c.Request.URL.Query().Get("date"))
	date -= date % 86400
	itms := []PartyItms{}
	err := DB.Select(&itms, `select item.description, quantity,
         rate from stktran left join item on itm_id=item.id where prt_id=? and
         lcn_id=? and type=? and date - (date % 86400)=?`, id, lc, tp, date)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, itms)
	}
}

func prtsumry(c *gin.Context) {
	id := c.Param("id")
	sumry := []PartySumry{}
	err := DB.Select(&sumry, `select (lcn_id || date || type) as id, lcn_id,
   location.description as locn, invoice, date, type, sum(value+tax) as amount
   from stktran left join location on lcn_id=location.id where prt_id=?
   group by date - (date % 86400), lcn_id, type order by date desc`, id)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, sumry)
	}
}

func partystk(c *gin.Context) {
	id := c.Param("id")
	dtfr, e1 := strconv.Atoi(c.Request.URL.Query().Get("dtfr"))
	if e1 != nil {
		dtfr = 0
	}
	dtto, e2 := strconv.Atoi(c.Request.URL.Query().Get("dtto"))
	if e2 != nil {
		dtto = 0
	}
	itmS := []PartyItms{}
	itmP := []PartyItms{}
	e1 = DB.Select(&itmS, `select type, date,
         item.description, quantity, rate from stktran 
         left join item on itm_id=item.id
         left join location on lcn_id=location.id
         where dummy=0 and prt_id=?
         and type='S' and date between ? and ?
         order by date desc`, id, dtfr, dtto)
	e2 = DB.Select(&itmP, `select type, date,
         item.description, quantity, rate from stktran 
         left join item on itm_id=item.id
         left join location on lcn_id=location.id
         where dummy=0 and prt_id=?
         and type='P' and date between ? and ?
         order by date desc`, id, dtfr, dtto)
	if e1 != nil || e2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"sale": e1, "purchase": e2})
		fmt.Println(e1, e2)
	} else {
		c.JSON(http.StatusOK, gin.H{"sale": itmS, "purchase": itmP})
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
		c.JSON(http.StatusBadRequest, err)
		fmt.Println(err)
	} else {
		c.JSON(http.StatusOK, pmts)
	}
}

func prtpayments(c *gin.Context) {
	id := c.Param("id")
	pmts := []PartyPmts{}
	err := DB.Select(&pmts, `select type, date,
         account.description as account, amount, comment
         from pmttran left join account on acc_id=account.id
         where prt_id=? order by date desc`, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	} else {
		c.JSON(http.StatusOK, pmts)
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
	c.JSON(http.StatusOK, pmts)
}
