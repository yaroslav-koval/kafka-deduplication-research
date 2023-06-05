package postgres

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type UseZeroInt int

func (uzi *UseZeroInt) ToInt() int {
	var res int

	if uzi != nil {
		res = int(*uzi)
	}

	return res
}

func NewUseZeroInt(v int) *UseZeroInt {
	vb := UseZeroInt(v)
	return &vb
}

func (uzi *UseZeroInt) Value() (driver.Value, error) {
	return driver.Value(uzi.ToInt()), nil
}

var _ driver.Valuer = (*UseZeroInt)(nil)

var _ sql.Scanner = (*UseZeroInt)(nil)

// Scan scans the time parsing it if necessary using timeFormat.
func (uzi *UseZeroInt) Scan(src interface{}) (err error) {
	switch src := src.(type) {
	case int:
		*uzi = UseZeroInt(src)
		return nil
	case int64:
		*uzi = UseZeroInt(int(src))
		return nil
	default:
		return fmt.Errorf("unsupported data type: %T", src)
	}
}
