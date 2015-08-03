package lidar

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

type PointRecord0 struct {
	PointData PointData
}

func (p PointRecord0) String() string {
	return getPrintString(&p)
}

type PointRecord1 struct {
	PointData PointData
	GPSTime   GPSTime
}

func (p PointRecord1) String() string {
	return getPrintString(&p)
}

type PointRecord2 struct {
	PointData PointData
	RGBData   RGBData
}

func (p PointRecord2) String() string {
	return getPrintString(&p)
}

type PointRecord3 struct {
	PointData PointData
	GPSTime   GPSTime
	RGBData   RGBData
}

func (p PointRecord3) String() string {
	return getPrintString(&p)
}

type PointData struct {
	X, Y, Z       int32
	Intensity     uint16
	BitField      PointBitField
	ClassField    ClassificationBitField
	ScanAngle     int8
	UserData      byte
	PointSourceID uint16
}

func (p PointData) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	s := reflect.ValueOf(&p).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		str := fmt.Sprintf("%s %s = %v\n", typeOfT.Field(i).Name, f.Type(), f.Interface())
		buffer.WriteString(str)
	}
	return buffer.String()
}

type PointBitField uint8

func (p *PointBitField) ReturnNumber() byte {
	//return byte((*p >> 4) & 0x07)
	return byte(*p & 7) //0x07)
}

func (p *PointBitField) NumberOfReturns() byte {
	//return byte((*p >> 3) & 0x07)
	return byte((*p >> 3) & 7)
}

func (p *PointBitField) ScanDirectionFlag() bool {
	//return bool(((*p >> 6) & 0x01) == 1)
	return bool(((*p >> 6) & 1) == 1)
}

func (p *PointBitField) EdgeOfFlightline() bool {
	//return bool(((*p >> 7) & 0x01) == 1)
	return bool(((*p >> 7) & 1) == 1)
}

func (p PointBitField) String() string {
	var buffer bytes.Buffer
	str := fmt.Sprintf("{\nRaw value binary byte = %v\n", strconv.FormatInt(int64(p), 2))
	buffer.WriteString(str)
	str = fmt.Sprintf("ReturnNumber byte = %v\n", p.ReturnNumber())
	buffer.WriteString(str)
	str = fmt.Sprintf("NumberOfReturns byte = %v\n", p.NumberOfReturns())
	buffer.WriteString(str)
	str = fmt.Sprintf("ScanDirection bool = %v\n", p.ScanDirectionFlag())
	buffer.WriteString(str)
	str = fmt.Sprintf("EdgeOfFlight bool = %v\n}", p.EdgeOfFlightline())
	buffer.WriteString(str)
	return buffer.String()
}

type ClassificationBitField uint8

func (c *ClassificationBitField) ClassValue() byte {
	//return byte((*c >> 3) & 0x1F)
	return byte(*c & 15)
}

func (c *ClassificationBitField) ClassString() string {
	cls := c.ClassValue()
	if m, ok := classMap[cls]; ok {
		return m
	}

	return "Undefined class value"
}

func (c *ClassificationBitField) IsSynthetic() bool {
	return bool(((*c >> 4) & 1) == 1)
}

func (c *ClassificationBitField) IsKeyPoint() bool {
	return bool(((*c >> 5) & 1) == 1)
}

func (c *ClassificationBitField) IsWithheld() bool {
	return bool(((*c >> 6) & 1) == 1)
}

func (c ClassificationBitField) String() string {
	var buffer bytes.Buffer
	str := fmt.Sprintf("{\nRaw value binary byte = %v\n", strconv.FormatInt(int64(c), 2))
	buffer.WriteString(str)
	str = fmt.Sprintf("ClassValue byte = %v\n", c.ClassValue())
	buffer.WriteString(str)
	str = fmt.Sprintf("ClassString string = %v\n", c.ClassString())
	buffer.WriteString(str)
	str = fmt.Sprintf("IsSynthetic bool = %v\n", c.IsSynthetic())
	buffer.WriteString(str)
	str = fmt.Sprintf("IsKeyPoint bool = %v\n", c.IsKeyPoint())
	buffer.WriteString(str)
	str = fmt.Sprintf("IsWithheld bool = %v\n}", c.IsWithheld())
	buffer.WriteString(str)
	return buffer.String()
}

var classMap = map[byte]string{
	0:  "Created, never classified",
	1:  "Unclassified1",
	2:  "Ground",
	3:  "Low Vegetation",
	4:  "Medium Vegetation",
	5:  "High Vegetation",
	6:  "Building",
	7:  "Low Point (noise)",
	8:  "Model Key-point (mass point)",
	9:  "Water",
	10: "Reserved for ASPRS Definition",
	11: "Reserved for ASPRS Definition",
	12: "Overlap Points2",
}

type GPSTime float64

type RGBData struct {
	Red, Green, Blue uint16
}

func getPrintString(p interface{}) string {
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	s := reflect.ValueOf(&p).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		str := fmt.Sprintf("%s %s = %v\n", typeOfT.Field(i).Name, f.Type(), f.Interface())
		buffer.WriteString(str)
	}
	return buffer.String()
}
