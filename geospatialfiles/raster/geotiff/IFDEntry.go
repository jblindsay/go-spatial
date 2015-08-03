package geotiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type IfdEntry struct {
	tag       GeoTiffTag
	dataType  GeotiffDataType
	count     uint32
	rawData   []byte
	byteOrder binary.ByteOrder
}

func (ifd *IfdEntry) AddData(data []byte) {
	if data != nil {
		ifd.rawData = append(ifd.rawData, data...)
	}
}

func CreateIfdEntry(code int, dataType GeotiffDataType, count uint32, data interface{}, byteOrder binary.ByteOrder) IfdEntry {
	var ret IfdEntry
	if myTag, ok := tagMap[code]; !ok {
		panic(errors.New("Unrecognized tag."))
	} else {
		ret.tag = myTag
	}
	ret.dataType = dataType
	ret.count = count
	ret.byteOrder = byteOrder
	if data != nil {
		if dataType != DT_ASCII {
			buf := new(bytes.Buffer)
			binary.Write(buf, byteOrder, data)

			ret.rawData = buf.Bytes()

		} else {
			if str, ok := data.(string); ok {
				ret.rawData = []byte(str)
			}
		}
	}

	return ret
}

func (ifd *IfdEntry) InterpretDataAsInt() (u []uint, err error) {
	u = make([]uint, ifd.count)
	switch ifd.dataType {
	case DT_Byte:
		for i := uint32(0); i < ifd.count; i++ {
			u[i] = uint(ifd.rawData[i])
		}
	case DT_Short:
		for i := uint32(0); i < ifd.count; i++ {
			u[i] = uint(ifd.byteOrder.Uint16(ifd.rawData[2*i : 2*(i+1)]))
		}
	case DT_Long:
		for i := uint32(0); i < ifd.count; i++ {
			u[i] = uint(ifd.byteOrder.Uint32(ifd.rawData[4*i : 4*(i+1)]))
		}
	default:
		return nil, UnsupportedDataTypeError
	}
	return u, nil
}

func (ifd *IfdEntry) InterpretDataAsFloat() (u []float64, err error) {
	u = make([]float64, ifd.count)
	switch ifd.dataType {
	case DT_Float:
		u2 := make([]float32, ifd.count)
		for i := uint32(0); i < ifd.count; i++ {
			// I'm not sure this code will work
			buf := bytes.NewReader(ifd.rawData[4*i : 4*(i+1)])
			binary.Read(buf, ifd.byteOrder, &u2[i])
		}
		for i := uint32(0); i < ifd.count; i++ {
			u[i] = float64(u2[i])
		}
	case DT_Double:
		for i := uint32(0); i < ifd.count; i++ {
			buf := bytes.NewReader(ifd.rawData[8*i : 8*(i+1)])
			binary.Read(buf, ifd.byteOrder, &u[i])
		}
	default:
		return nil, UnsupportedDataTypeError
	}
	return u, nil
}

func (ifd *IfdEntry) InterpretDataAsRational() (u []float64, err error) {
	u = make([]float64, ifd.count)
	switch ifd.dataType {
	case DT_Rational:
		offset := 0
		for i := uint32(0); i < ifd.count; i++ {
			v1 := uint(ifd.byteOrder.Uint32(ifd.rawData[offset : offset+4]))
			v2 := uint(ifd.byteOrder.Uint32(ifd.rawData[offset+4 : offset+8]))
			u[i] = float64(v1) / float64(v2)
			offset += 8
		}
	default:
		return nil, UnsupportedDataTypeError
	}
	return u, nil
}

// ifdFloat decodes the IFD entry in p, which must be of the ASCII
// type, and returns the decoded uint values.
func (ifd *IfdEntry) InterpretDataAsASCII() (u []string, err error) {
	u = make([]string, 1)
	switch ifd.dataType {
	case DT_ASCII:
		u[0] = string(ifd.rawData[:ifd.count-1])
	default:
		return nil, UnsupportedDataTypeError
	}
	return u, nil
}

func (ifd IfdEntry) String() string {
	s := ifd.tag
	retVal := fmt.Sprintf("%v , DataType: %v, Count: %v", s, ifd.dataType, ifd.count)
	switch ifd.dataType {
	case DT_Byte,
		DT_Long:
		v, _ := ifd.InterpretDataAsInt()
		return fmt.Sprintf("%s Value: %v", retVal, v)
	case DT_Short:
		v, _ := ifd.InterpretDataAsInt()
		if ifd.count == 1 {
			if strVal, ok := tagLookupTable(&ifd); ok == nil {
				return fmt.Sprintf("%s Value: [%v, %s]", retVal, v[0], strVal)
			} else {
				return fmt.Sprintf("%s Value: %v", retVal, v)
			}
		} else {
			return fmt.Sprintf("%s Value: %v", retVal, v)
		}
	case DT_Rational:
		v, _ := ifd.InterpretDataAsRational()
		return fmt.Sprintf("%s Value: %v", retVal, v)
	case DT_Float,
		DT_Double:
		v, _ := ifd.InterpretDataAsFloat()
		return fmt.Sprintf("%s Value: %v", retVal, v)

	case DT_ASCII:
		v, _ := ifd.InterpretDataAsASCII()
		return fmt.Sprintf("%s Value: %v", retVal, v)
	}
	return retVal
}

// make a slice of IfdEntries sortable by its GeoTiffTag code
type ifdSortedByCode []IfdEntry

func (a ifdSortedByCode) Len() int           { return len(a) }
func (a ifdSortedByCode) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ifdSortedByCode) Less(i, j int) bool { return a[i].tag.Code < a[j].tag.Code }
