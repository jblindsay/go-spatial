// Copyright 2014 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Nov. 2014.

// Package raster provides support for reading and creating various common
// geospatial raster data formats.
package raster

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"

	"path/filepath"
	"strings"
)

type rasterData interface {
	InitializeRaster(fileName string,
		rows int, columns int, north float64, south float64,
		east float64, west float64, config *RasterConfig) error
	FileName() string
	SetFileName(value string) error
	Rows() int
	Columns() int
	SetRows(value int)
	SetColumns(value int)
	North() float64
	South() float64
	East() float64
	West() float64
	MinimumValue() float64
	MaximumValue() float64
	RasterType() RasterType
	NoData() float64
	SetNoData(value float64)
	ByteOrder() binary.ByteOrder
	SetByteOrder(value binary.ByteOrder)
	Value(index int) float64
	SetValue(index int, value float64)
	Data() ([]float64, error)
	SetData(values []float64)
	Save() error
	MetadataEntries() []string
	AddMetadataEntry(value string)
	SetRasterConfig(value *RasterConfig)
	GetRasterConfig() *RasterConfig
}

type Raster struct {
	Rows, Columns            int
	NumberofCells            int
	North, South, East, West float64
	NoDataValue              float64
	FileName                 string
	FileExtension            string
	RasterFormat             RasterType
	ByteOrder                binary.ByteOrder
	rd                       rasterData
	reflectAtBoundaries      bool
}

type RasterConfig struct {
	NoDataValue               float64
	InitialValue              float64
	RasterFormat              RasterType
	ByteOrder                 binary.ByteOrder
	MetadataEntries           []string
	CoordinateRefSystemWKT    string
	NumberOfBands             int
	PhotometricInterpretation int
	DataType                  int
	PaletteNonlinearity       float64
	ZUnits                    string
	XYUnits                   string
	PreferredPalette          string
	DisplayMinimum            float64
	DisplayMaximum            float64
	ReflectAtBoundaries       bool
	PixelIsArea               bool
	EPSGCode                  int
}

func (h RasterConfig) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Raster Configuration:\n")
	s := reflect.ValueOf(&h).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		str := fmt.Sprintf("%s %s = %v\n", typeOfT.Field(i).Name, f.Type(), f.Interface())
		buffer.WriteString(str)
	}
	return buffer.String()
}

func NewDefaultRasterConfig() *RasterConfig {
	var rc RasterConfig
	rc.NoDataValue = -32768.0
	rc.InitialValue = -32768.0
	rc.RasterFormat = RT_UnknownRaster
	rc.ByteOrder = binary.LittleEndian
	rc.NumberOfBands = 1
	rc.PaletteNonlinearity = 1.0
	rc.ZUnits = "not specified"
	rc.XYUnits = "not specified"
	rc.PreferredPalette = "not specified"
	rc.DisplayMinimum = math.MaxFloat64
	rc.DisplayMaximum = -math.MaxFloat64
	rc.CoordinateRefSystemWKT = ""
	rc.PixelIsArea = true
	rc.PhotometricInterpretation = -1
	rc.DataType = -1
	rc.MetadataEntries = make([]string, 1)
	return &rc
}

// Data Type
const (
	DT_INT8 = iota
	DT_UINT8
	DT_INT16
	DT_UINT16
	DT_INT32
	DT_UINT32
	DT_INT64
	DT_UINT64
	DT_FLOAT32
	DT_FLOAT64
	DT_RGB24
	DT_RGB48
	DT_RGBA32
	DT_RGBA64
	DT_PALETTED
)

func CreateNewRaster(fileName string, rows int, columns int, north float64,
	south float64, east float64, west float64, config ...*RasterConfig) (*Raster, error) {

	var err error
	var myConfig *RasterConfig
	if len(config) == 0 {
		myConfig = NewDefaultRasterConfig()
	} else {
		// obviously, more than one config could be specified because of
		// the use of the variadic parameter, which is done so that it
		// is possible to specify no config. If more than one config is
		// specified, only the last is used.
		myConfig = config[len(config)-1]
	}
	var r Raster
	var rasterType RasterType
	if myConfig.RasterFormat != RT_UnknownRaster {
		rasterType = myConfig.RasterFormat
	} else {
		rasterType, err = DetermineRasterFormat(fileName)
		if err == UnsupportedRasterFormatError {
			return &r, err
		}
	}
	r.RasterFormat = rasterType

	var myRasterData rasterData

	switch rasterType {
	case RT_ArcGisBinaryRaster:
		myRasterData = new(arcGisBinaryRaster)

	case RT_ArcGisAsciiRaster:
		myRasterData = new(arcGisAsciiRaster)

	case RT_WhiteboxRaster:
		myRasterData = new(whiteboxRaster)

	case RT_GrassAsciiRaster:
		myRasterData = new(grassAsciiRaster)

	case RT_GeoTiff:
		myRasterData = new(geotiffRaster)

	case RT_IdrisiRaster:
		myRasterData = new(idrisiRaster)

	}

	r.reflectAtBoundaries = myConfig.ReflectAtBoundaries

	err = myRasterData.InitializeRaster(fileName, rows, columns, north, south, east, west, myConfig)
	if err != nil {
		return &r, RasterInitializationError
	}
	r.rd = myRasterData
	setVariablesFromRasterData(&r, r.rd)

	return &r, nil
}

func CreateRasterFromFile(fileName string, config ...RasterConfig) (*Raster, error) {
	var r Raster
	var err error
	r.FileName = fileName
	r.FileExtension = strings.ToLower(filepath.Ext(r.FileName))

	// what is the raster format?
	var rt RasterType
	if len(config) > 0 {
		// if the config is specified, see if you can use it's
		// RasterFormat for the RasterType, if not, try to determine
		// what it is.
		// obviously, more than one config could be specified because of
		// the use of the variadic parameter, which is done so that it
		// is possible to specify no config. If more than one config is
		// specified, only the last is used.
		rt = config[len(config)-1].RasterFormat
		if rt == RT_UnknownRaster {
			rt, err = DetermineRasterFormat(fileName)
			if err != nil || rt == RT_UnknownRaster {
				return &r, err
			}
		}
	} else {
		rt, err = DetermineRasterFormat(fileName)
		if err != nil || rt == RT_UnknownRaster {
			return &r, err
		}
	}
	r.RasterFormat = rt

	// see if it is a supported raster format
	//if !IsSupportedRasterFileExtension(fileName) {
	//	return &r, fmt.Errorf(`Unsupported raster format: "%s"`, r.FileExtension)
	//}

	r.rd, err = r.getRasterData()
	r.check(err)
	if r.rd == nil {
		return &r, RasterInitializationError
	}

	setVariablesFromRasterData(&r, r.rd)

	return &r, nil

}

func (r *Raster) getRasterData() (rasterData, error) {

	switch r.RasterFormat {
	case RT_GeoTiff:
		myGeoTiff := new(geotiffRaster)
		myGeoTiff.SetFileName(r.FileName)
		return myGeoTiff, nil

	case RT_ArcGisBinaryRaster:
		myArcRaster := new(arcGisBinaryRaster)
		myArcRaster.SetFileName(r.FileName)
		return myArcRaster, nil

	case RT_ArcGisAsciiRaster:
		myArcRaster := new(arcGisAsciiRaster)
		myArcRaster.SetFileName(r.FileName)
		return myArcRaster, nil

	case RT_WhiteboxRaster:
		myWhiteboxRaster := new(whiteboxRaster)
		myWhiteboxRaster.SetFileName(r.FileName)
		return myWhiteboxRaster, nil

	case RT_GrassAsciiRaster:
		myGrassRaster := new(grassAsciiRaster)
		myGrassRaster.SetFileName(r.FileName)
		return myGrassRaster, nil

	case RT_IdrisiRaster:
		myIdrisiRaster := new(idrisiRaster)
		myIdrisiRaster.SetFileName(r.FileName)
		return myIdrisiRaster, nil
	}

	return nil, nil
}

// Retrives an individual pixel value in the grid.
func (r *Raster) Value(row, column int) float64 {
	if column >= 0 && column < r.Columns && row >= 0 && row < r.Rows {
		// what is the cell number?
		cellNum := row*r.Columns + column
		return r.rd.Value(cellNum)
	} else {
		if !r.reflectAtBoundaries {
			return r.rd.NoData()
		}

		// if you get to this point, it is reflected at the edges
		if row < 0 {
			row = -row - 1
		}
		if row >= r.Rows {
			row = r.Rows - (row - r.Rows) - 1
		}
		if column < 0 {
			column = -column - 1
		}
		if column >= r.Columns {
			column = r.Columns - (column - r.Columns) - 1
		}
		if column >= 0 && column < r.Columns && row >= 0 && row < r.Rows {
			return r.Value(row, column)
		} else {
			// it was too off grid to be reflected.
			return r.rd.NoData()
		}
	}
}

// Sets an individual pixel value in the grid.
func (r *Raster) SetValue(row, column int, value float64) {
	if column >= 0 && column < r.Columns && row >= 0 && row < r.Rows {
		// what is the cell number?
		cellNum := row*r.Columns + column
		r.rd.SetValue(cellNum, value)
	}
}

// Returns the data as a slice of float64 values
func (r *Raster) Data() ([]float64, error) {
	return r.rd.Data()
}

// Sets the data from a slice of float64 values
func (r *Raster) SetData(values []float64) {
	r.rd.SetData(values)
}

func (r *Raster) Save() (err error) {
	return r.rd.Save()
}

// Sets the raster config
func (r *Raster) SetRasterConfig(value *RasterConfig) {
	r.rd.SetRasterConfig(value)
	r.reflectAtBoundaries = value.ReflectAtBoundaries
}

// Gets the raster config
func (r *Raster) GetRasterConfig() *RasterConfig {
	return r.rd.GetRasterConfig()
}

func (r *Raster) GetMetadataEntries() []string {
	return r.rd.MetadataEntries()
}

func (r *Raster) AddMetadataEntry(value string) {
	r.rd.AddMetadataEntry(value)
}

func (r *Raster) GetMinimumValue() float64 {
	return r.rd.MinimumValue()
}

func (r *Raster) GetMaximumValue() float64 {
	return r.rd.MaximumValue()
}

func (r *Raster) GetCellSizeX() (cellSizeX float64) {
	if r.rd.GetRasterConfig().PixelIsArea {
		cellSizeX = (r.East - r.West) / (float64(r.Columns))
	} else {
		cellSizeX = (r.East - r.West) / (float64(r.Columns - 1))
	}
	return cellSizeX
}

func (r *Raster) GetCellSizeY() (cellSizeY float64) {
	if r.rd.GetRasterConfig().PixelIsArea {
		cellSizeY = (r.North - r.South) / (float64(r.Rows))
	} else {
		cellSizeY = (r.North - r.South) / (float64(r.Rows - 1))
	}
	return cellSizeY
}

func (r *Raster) check(e error) {
	if e != nil {
		panic(e)
	}
}

// set's the Raster's public variables based on a RasterData
func setVariablesFromRasterData(r *Raster, rd rasterData) (err error) {
	r.Columns = rd.Columns()
	r.Rows = rd.Rows()
	r.North = rd.North()
	r.South = rd.South()
	r.East = rd.East()
	r.West = rd.West()
	r.ByteOrder = rd.ByteOrder()
	r.NoDataValue = rd.NoData()
	r.NumberofCells = r.Rows * r.Columns
	return nil
}
