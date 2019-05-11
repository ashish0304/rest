package main

type PartyAcc struct {
	Id          int     `db:"id" json:"id"`
	Description string  `db:"description" json:"description"`
	Balance     float32 `db:"balance" json:"balance"`
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
	Acc_id  NullInt64  `json:"acc_id" db:"acc_id"`
	Account NullString `json:"account" db:"account"`
	Comment NullString `json:"comment" db:"comment"`
	Amount  float32    `json:"amount" db:"amount"`
}

type Locnstat struct {
	Type   string      `json:"type" db:"type"`
	Amount float32     `json:"amount" db:"amount"`
	Tax    NullFloat64 `json:"tax" db:"tax"`
}

type PartyItms struct {
	Type     NullString `json:"type" db:"type"`
	Date     NullInt64  `db:"date" json:"date"`
	Item     string     `db:"description" json:"description"`
	Quantity int32      `db:"quantity" json:"quantity"`
	Rate     float32    `db:"rate" json:"rate"`
}

type PartySumry struct {
	Id      string     `db:"id" json:"id"`
	Lcn_id  NullInt64  `json:"lcn_id" db:"lcn_id"`
	Locn    NullString `db:"locn" json:"locn"`
	Type    string     `db:"type" json:"type"`
	Invoice NullString `db:"invoice" json:"invoice"`
	Date    int64      `db:"date" json:"date"`
	Amount  float32    `db:"amount" json:"amount"`
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

type GSTReps struct {
	Type string `db:"type" json:"type"`
	Date int64 `db:"date" json:"date"`
	Invoice NullString `db:"invoice" json:"invoice"`
	Party NullString `db:"party" json:"party"`
	GSTN NullString `db:"gstn" json:"gstn"`
	HSN NullString `db:"hsn" json:"hsn"`
	TRate NullFloat64 `db:"trate" json:"trate"`
	Amount float32 `db:"amount" json:"amount"`
	Tax NullFloat64 `db:"tax" json:"tax"`
}
