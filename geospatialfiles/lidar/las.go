package lidar

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

var println = fmt.Println
var bo = binary.LittleEndian

type LasFile struct {
	fileName    string
	r           *os.File // io.ReaderAt
	pointData   []PointData
	gpsTimeData []GPSTime
	Header      LasHeader
}

type VlrHeader struct { // header of the variable length record (VLR)
	UserID                  string // 16 characters
	RecordID                uint16
	RecordLengthAfterHeader uint16
	Description             string // 32 characters
}

type ExtendedVlrHeader struct { // header of the extended variable length record (EVLR)
	UserID                  string // 16 characters
	RecordID                uint16
	RecordLengthAfterHeader uint64
	Description             string // 32 characters
}

func CreateFromFile(fileName string) (*LasFile, error) {
	var las LasFile
	las.fileName = fileName
	las.readFile()
	return &las, nil
}

func (las LasFile) GetPointTypeInfo() string {
	switch las.Header.PointFormatID {
	//case 0:
	//	var p PointRecord0
	//	return p.String()
	case 1:
		//return getPrintString(PointRecord1{})
	default:
		//return getPrintString(PointRecord0{})
	}

	return ""
}

func (las *LasFile) readFile() {
	// open the file
	r, err := os.Open(las.fileName)
	if err != nil {
		panic(err)
	}
	las.r = r
	las.readHeader()
}

func (las *LasFile) readPointData() {
	las.pointData = make([]PointData, las.Header.NumberPoints)
	pointRecLen := int64(las.Header.PointRecordLength)
	initialOffset := int64(las.Header.OffsetToPoints)
	for i := int64(0); i < int64(las.Header.NumberPoints); i++ {
		offset := initialOffset + pointRecLen*i
		las.r.Seek(offset, 0)
		if err := binary.Read(las.r, bo, &las.pointData[i]); err != nil {
			panic(errors.New("Error reading point data"))
		}
	}

}

func (las *LasFile) GetFileName() string {
	return las.fileName
}

func (las *LasFile) GetPointXYZ(n int64) (X, Y, Z float64) {
	if las.pointData == nil {
		las.readPointData()
	}
	X = float64(las.pointData[n].X)*las.Header.XScaleFactor + las.Header.XOffset
	Y = float64(las.pointData[n].Y)*las.Header.YScaleFactor + las.Header.YOffset
	Z = float64(las.pointData[n].Z)*las.Header.ZScaleFactor + las.Header.ZOffset
	return X, Y, Z
}

func (las *LasFile) GetPointIntensity(n int64) uint16 {
	if las.pointData == nil {
		las.readPointData()
	}
	return las.pointData[n].Intensity
}

func (las *LasFile) GetPointClassValue(n int64) byte {
	if las.pointData == nil {
		las.readPointData()
	}
	return las.pointData[n].ClassField.ClassValue()
}

func (las *LasFile) GetPointClassName(n int64) string {
	if las.pointData == nil {
		las.readPointData()
	}
	return las.pointData[n].ClassField.ClassString()
}

func (las *LasFile) PrintPointData(n int64) {
	if las.pointData == nil {
		las.readPointData()
	}
	println(las.pointData[n].String())
}

func (las *LasFile) Close() error {
	return las.r.Close()
}
