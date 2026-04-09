package athena

import (
	"github.com/rilldata/rill/runtime/drivers"
)

type dialect struct {
	drivers.BaseDialect
}

var DialectAthena drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.InitBase(d)
	return d
}()

func (d *dialect) String() string { return drivers.DialectNameAthena }
