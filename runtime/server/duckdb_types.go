package server

var INTEGERS = map[string]bool{
	"BIGINT":   true,
	"HUGEINT":  true,
	"SMALLINT": true,
	"INTEGER":  true,
	"TINYINT":  true,
	"UBIGINT":  true,
	"UINTEGER": true,
	"UTINYINT": true,
	"INT1":     true,
	"INT4":     true,
	"INT":      true,
	"SIGNED":   true,
	"SHORT":    true,
}

var FLOATS = map[string]bool{
	"DOUBLE":  true,
	"DECIMAL": true,
	"FLOAT8":  true,
	"NUMERIC": true,
	"FLOAT":   true,
}

var DATES = map[string]bool{
	"DATE": true,
}

// NUMERICS It's a copy of INTEGERS and FLOATS map until we have a better way to reference the maps directly
var NUMERICS = map[string]bool{
	"BIGINT":   true,
	"HUGEINT":  true,
	"SMALLINT": true,
	"INTEGER":  true,
	"TINYINT":  true,
	"UBIGINT":  true,
	"UINTEGER": true,
	"UTINYINT": true,
	"INT1":     true,
	"INT4":     true,
	"INT":      true,
	"SIGNED":   true,
	"SHORT":    true,
	"DOUBLE":   true,
	"DECIMAL":  true,
	"FLOAT8":   true,
	"NUMERIC":  true,
	"FLOAT":    true,
}

var BOOLEANS = map[string]bool{
	"BOOLEAN": true,
	"BOOL":    true,
	"LOGICAL": true,
}

// TIMESTAMPS and DATES map fields
var TIMESTAMPS = map[string]bool{
	"TIMESTAMP": true,
	"TIME":      true,
	"DATETIME":  true,
	"DATE":      true,
}

var INTERVALS = map[string]bool{
	"INTERVAL": true,
}

var STRING_LIKES = map[string]bool{
	"BYTE_ARRAY": true,
	"VARCHAR":    true,
	"CHAR":       true,
	"BPCHAR":     true,
	"TEXT":       true,
	"STRING":     true,
}

// copy of STRING_LIKES and BOOLEANS map
var CATEGORICALS = map[string]bool{
	"BOOLEAN":    true,
	"BOOL":       true,
	"LOGICAL":    true,
	"BYTE_ARRAY": true,
	"VARCHAR":    true,
	"CHAR":       true,
	"BPCHAR":     true,
	"TEXT":       true,
	"STRING":     true,
}

var ANY_TYPES = map[string]bool{
	"BIGINT":     true,
	"HUGEINT":    true,
	"SMALLINT":   true,
	"INTEGER":    true,
	"TINYINT":    true,
	"UBIGINT":    true,
	"UINTEGER":   true,
	"UTINYINT":   true,
	"INT1":       true,
	"INT4":       true,
	"INT":        true,
	"SIGNED":     true,
	"SHORT":      true,
	"DOUBLE":     true,
	"DECIMAL":    true,
	"FLOAT8":     true,
	"NUMERIC":    true,
	"FLOAT":      true,
	"BOOLEAN":    true,
	"BOOL":       true,
	"LOGICAL":    true,
	"TIMESTAMP":  true,
	"TIME":       true,
	"DATETIME":   true,
	"DATE":       true,
	"INTERVAL":   true,
	"BYTE_ARRAY": true,
	"VARCHAR":    true,
	"CHAR":       true,
	"BPCHAR":     true,
	"TEXT":       true,
	"STRING":     true,
}

var TypesMap = map[string]map[string]bool{
	"INTEGERS":     INTEGERS,
	"FLOATS":       FLOATS,
	"NUMERICS":     NUMERICS,
	"BOOLEANS":     BOOLEANS,
	"TIMESTAMPS":   TIMESTAMPS,
	"INTERVALS":    INTERVALS,
	"CATEGORICALS": CATEGORICALS,
	"ANY_TYPES":    ANY_TYPES,
}
