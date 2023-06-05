package postgres

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type UseZeroBool bool

func (uzb *UseZeroBool) ToBool() bool {
	var res bool

	if uzb != nil {
		res = bool(*uzb)
	}

	return res
}

func NewUseZeroBool(v bool) *UseZeroBool {
	vb := UseZeroBool(v)
	return &vb
}

func (uzb *UseZeroBool) Value() (driver.Value, error) {
	return driver.Value(uzb.ToBool()), nil
}

var _ driver.Valuer = (*UseZeroBool)(nil)

var _ sql.Scanner = (*UseZeroBool)(nil)

// Scan scans the time parsing it if necessary using timeFormat.
func (uzb *UseZeroBool) Scan(src interface{}) (err error) {
	switch src := src.(type) {
	case bool:
		*uzb = UseZeroBool(src)
		return nil
	default:
		return fmt.Errorf("unsupported data type: %T", src)
	}
}
