package geotiff

type GeotiffDataType int

// Data types (p. 14-16 of the spec).
const (
	DT_Byte      = 1
	DT_ASCII     = 2
	DT_Short     = 3
	DT_Long      = 4
	DT_Rational  = 5
	DT_Sbyte     = 6
	DT_Undefined = 7
	DT_Sshort    = 8
	DT_Slong     = 9
	DT_Srational = 10
	DT_Float     = 11
	DT_Double    = 12
)

// The length of one instance of each data type in bytes.
var dataTypeLengths = [...]uint32{0, 1, 1, 2, 4, 8, 1, 2, 2, 4, 8, 8, 16}

var dataTypeList = []string{
	"Byte",
	"ASCII",
	"Short",
	"Long",
	"Rational",
	"Sbyte",
	"Undefined",
	"Sshort",
	"Slong",
	"Srational",
	"Float",
	"Double",
}

// String returns the English name of the DataType ("Byte", "ASCII", ...).
func (g GeotiffDataType) String() string { return dataTypeList[g-1] }

func (g GeotiffDataType) GetBitLength() uint32 {
	return dataTypeLengths[g]
}
