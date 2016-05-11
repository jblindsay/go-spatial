package lidar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
)

type LasHeader struct {
	FileSignature        string //[4]byte
	FileSourceID         uint16
	GlobalEncoding       uint16
	ProjectID1           uint32
	ProjectID2           uint16
	ProjectID3           uint16
	ProjectID4           uint64
	VersionMajor         byte
	VersionMinor         byte
	SystemID             string // 32 characters
	GeneratingSoftware   string // 32 characters
	FileCreationDay      uint16
	FileCreationYear     uint16
	HeaderSize           uint16
	OffsetToPoints       uint32
	NumberOfVLRs         uint32
	PointFormatID        byte
	PointRecordLength    uint16
	NumberPoints         uint32
	NumberPointsByReturn [7]uint32
	XScaleFactor         float64
	YScaleFactor         float64
	ZScaleFactor         float64
	XOffset              float64
	YOffset              float64
	ZOffset              float64
	MaxX                 float64
	MinX                 float64
	MaxY                 float64
	MinY                 float64
	MaxZ                 float64
	MinZ                 float64
	WaveformDataStart    uint64
}

func (las *LasFile) readHeader() {

	b := make([]byte, 243)
	if _, err := las.r.ReadAt(b[0:243], 0); err != nil && err != io.EOF {
		panic(err)
	}

	las.Header.FileSignature = string(b[0:4])
	las.Header.FileSourceID = binary.LittleEndian.Uint16(b[4:6])
	las.Header.GlobalEncoding = binary.LittleEndian.Uint16(b[6:8])
	las.Header.ProjectID1 = binary.LittleEndian.Uint32(b[8:12])
	las.Header.ProjectID2 = binary.LittleEndian.Uint16(b[12:14])
	las.Header.ProjectID3 = binary.LittleEndian.Uint16(b[14:16])
	las.Header.ProjectID4 = binary.LittleEndian.Uint64(b[16:24])
	las.Header.VersionMajor = b[24]
	las.Header.VersionMinor = b[25]
	las.Header.SystemID = string(b[26:58])
	las.Header.SystemID = strings.Trim(las.Header.SystemID, " ")
	las.Header.GeneratingSoftware = string(b[58:90])
	las.Header.GeneratingSoftware = strings.Trim(las.Header.GeneratingSoftware, " ")
	las.Header.FileCreationDay = binary.LittleEndian.Uint16(b[90:92])
	las.Header.FileCreationYear = binary.LittleEndian.Uint16(b[92:94])
	las.Header.HeaderSize = binary.LittleEndian.Uint16(b[94:96])
	las.Header.OffsetToPoints = binary.LittleEndian.Uint32(b[96:100])
	las.Header.NumberOfVLRs = binary.LittleEndian.Uint32(b[100:104])
	las.Header.PointFormatID = b[104]
	las.Header.PointRecordLength = binary.LittleEndian.Uint16(b[105:107])
	las.Header.NumberPoints = binary.LittleEndian.Uint32(b[107:111])

	offset := 111
	var numReturns int
	if las.Header.VersionMajor == 1 && (las.Header.VersionMinor <= 3) {
		numReturns = 5
	} else if las.Header.VersionMajor == 1 && las.Header.VersionMinor > 3 {
		numReturns = 7
	} else {
		panic(errors.New("Unsupported LAS file type"))
	}

	for i := 0; i < numReturns; i++ {
		las.Header.NumberPointsByReturn[i] = binary.LittleEndian.Uint32(b[offset : offset+4])
		offset += 4
	}

	las.Header.XScaleFactor = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.YScaleFactor = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.ZScaleFactor = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.XOffset = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.YOffset = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.ZOffset = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.MaxX = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.MinX = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.MaxY = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.MinY = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.MaxZ = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	las.Header.MinZ = math.Float64frombits(binary.LittleEndian.Uint64(b[offset : offset+8]))
	offset += 8
	if las.Header.VersionMajor == 1 && las.Header.VersionMinor == 3 {
		las.Header.WaveformDataStart = binary.LittleEndian.Uint64(b[offset : offset+8])
	}
}

func (h LasHeader) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("LAS File Header:\n")
	s := reflect.ValueOf(&h).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		str := fmt.Sprintf("%s %s = %v\n", typeOfT.Field(i).Name, f.Type(), f.Interface())
		buffer.WriteString(str)
	}
	return buffer.String()
}
