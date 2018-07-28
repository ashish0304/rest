package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
)

func NullZero(z uint32) interface{} {
	if z == 0 {
		return nil
	} else {
		return z
	}
}

type NullInt64 sql.NullInt64

func (ni *NullInt64) Scan(value interface{}) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ni = NullInt64{i.Int64, false}
	} else {
		*ni = NullInt64{i.Int64, true}
	}
	return nil
}

func (p NullInt64) Value() (driver.Value, error) {
	if !p.Valid {
		return driver.Value(nil), nil
	}
	return driver.Value(p.Int64), nil
}

// NullBool is an alias for sql.NullBool data type
type NullBool sql.NullBool

// Scan implements the Scanner interface for NullBool
func (nb *NullBool) Scan(value interface{}) error {
	//fmt.Printf("%#v",&nb)
	var b sql.NullBool
	if err := b.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*nb = NullBool{b.Bool, false}
	} else {
		*nb = NullBool{b.Bool, true}
	}

	return nil
}

func (p NullBool) Value() (driver.Value, error) {
	if !p.Valid {
		return driver.Value(nil), nil
	}
	return driver.Value(p.Bool), nil
}

// NullFloat64 is an alias for sql.NullFloat64 data type
type NullFloat64 sql.NullFloat64

// Scan implements the Scanner interface for NullFloat64
func (nf *NullFloat64) Scan(value interface{}) error {
	var f sql.NullFloat64
	if err := f.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*nf = NullFloat64{f.Float64, false}
	} else {
		*nf = NullFloat64{f.Float64, true}
	}

	return nil
}

func (p NullFloat64) Value() (driver.Value, error) {
	if !p.Valid {
		return driver.Value(nil), nil
	}
	return driver.Value(p.Float64), nil
}

// NullString is an alias for sql.NullString data type
type NullString sql.NullString

// Scan implements the Scanner interface for NullString
func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ns = NullString{s.String, false}
	} else {
		*ns = NullString{s.String, true}
	}

	return nil
}

func (p NullString) Value() (driver.Value, error) {
	if !p.Valid {
		return driver.Value(nil), nil
	}
	return driver.Value(p.String), nil
}

// MarshalJSON for NullInt64
func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

// UnmarshalJSON for NullInt64
func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ni.Int64)
	ni.Valid = (err == nil)
	return err
}

// MarshalJSON for NullBool
func (nb NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nb.Bool)
}

// UnmarshalJSON for NullBool
func (nb *NullBool) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &nb.Bool)
	nb.Valid = (err == nil)
	return err
}

// MarshalJSON for NullFloat64
func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nf.Float64)
}

// UnmarshalJSON for NullFloat64
func (nf *NullFloat64) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &nf.Float64)
	nf.Valid = (err == nil)
	return err
}

// MarshalJSON for NullString
func (ns *NullString) xMarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON for NullString
func (ns *NullString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ns.String)
	ns.Valid = (err == nil)
	return err
}

func (ns NullString) xUnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ns.String)
	ns.Valid = (err == nil)
	return err
}
