package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func gstreps(c *gin.Context) {
	repSale := [] GSTReps{}
	repPurc := [] GSTReps{}
	dtfr := c.Request.URL.Query().Get("dtfr")
	dtto := c.Request.URL.Query().Get("dtto")
	locn := c.Request.URL.Query().Get("locn")
	
	err1 := DB.Select(&repSale, `select type, date, invoice,
	        party.description as party, gstn, item.hsn,
            round((stktran.tax/(value/100)),1) as trate, value as amount,
	        stktran.tax from stktran left join party on stktran.prt_id=party.id
            left join item on itm_id=item.id
            where type='S' and lcn_id=? and date between ? and ?
            order by date`, locn, dtfr, dtto)
           
    err2 := DB.Select(&repPurc, `select type, date, invoice,
	        party.description as party, gstn, item.hsn,
            round((stktran.tax/(value/100)),1) as trate, value as amount,
	        stktran.tax from stktran left join party on stktran.prt_id=party.id
	        left join item on itm_id=item.id
            where type='P' and lcn_id=? and date between ? and ?
            order by date`, locn, dtfr, dtto)
           
	if err1 != nil || err2 != nil{
		fmt.Println(err1, err2)
	}else{
	    c.JSON(http.StatusOK, gin.H{"sales": repSale, "purchase": repPurc})
	}
}


