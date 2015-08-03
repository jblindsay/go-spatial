// Copyright 2014 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Nov. 2014.

// Package raster provides support for reading and creating various common
// geospatial raster data formats.
package raster

import (
	"encoding/binary"
	"errors"
	"gospatial/geospatialfiles/raster/geotiff"
	"math"
	"os"
	"strconv"
	"strings"
)

// Used to manipulate an ArcGIS ASCII raster file.
type geotiffRaster struct {
	fileName     string
	data         []float64
	header       geotiffRasterHeader
	minimumValue float64
	maximumValue float64
	config       *RasterConfig
	gt           geotiff.GeoTIFF
}

func (r *geotiffRaster) InitializeRaster(fileName string,
	rows int, columns int, north float64, south float64,
	east float64, west float64, config *RasterConfig) (err error) {

	r.config = config
	// set the various rows, columns, north, etc.
	r.header.columns = columns
	r.header.rows = rows
	r.header.numCells = rows * columns
	r.header.north = north
	r.header.south = south
	r.header.east = east
	r.header.west = west
	r.header.cellSize = (east - west) / float64(r.header.columns)
	//r.header.nodata = config.NoDataValue

	r.fileName = fileName

	// does the file already exist? If yes, delete it.
	if _, err = os.Stat(r.fileName); err == nil {
		if err = os.Remove(r.fileName); err != nil {
			return FileDeletingError
		}
	}

	// initialize the data array
	r.data = make([]float64, r.header.numCells)
	if config.InitialValue != 0 {
		for i := range r.data {
			r.data[i] = config.InitialValue
		}
	}

	r.minimumValue = math.MaxFloat64
	r.maximumValue = -math.MaxFloat64

	var bitsPerSample []uint
	switch r.config.DataType {
	default:
		bitsPerSample = []uint{32}
	case DT_FLOAT64, DT_UINT64, DT_INT64:
		bitsPerSample = []uint{64}

	case DT_UINT16, DT_INT16:
		bitsPerSample = []uint{16}

	case DT_UINT8, DT_INT8:
		bitsPerSample = []uint{8}

	case DT_RGB24:
		bitsPerSample = []uint{8, 8, 8}

	case DT_RGB48:
		bitsPerSample = []uint{16, 16, 16}

	case DT_RGBA32:
		bitsPerSample = []uint{8, 8, 8, 8}

	case DT_RGBA64:
		bitsPerSample = []uint{16, 16, 16, 16}
	}

	var sampleFormat uint
	switch r.config.DataType {

	case DT_INT8, DT_INT16, DT_INT32, DT_INT64:
		sampleFormat = geotiff.SF_SignedInteger

	case DT_FLOAT32, DT_FLOAT64:
		sampleFormat = geotiff.SF_FloatingPoint

	default:
		sampleFormat = geotiff.SF_UnsignedInteger
	}

	if r.config.PhotometricInterpretation < 1 {
		switch r.config.DataType {
		case DT_RGB24, DT_RGB48, DT_RGBA32, DT_RGBA64:
			r.config.PhotometricInterpretation = geotiff.PI_RGB

		case DT_PALETTED:
			r.config.PhotometricInterpretation = geotiff.PI_Paletted

		default:
			r.config.PhotometricInterpretation = geotiff.PI_BlackIsZero
		}
	}

	r.gt = geotiff.GeoTIFF{Rows: uint(rows), Columns: uint(columns),
		ByteOrder: r.config.ByteOrder, BitsPerSample: bitsPerSample,
		SampleFormat: sampleFormat, PhotometricInterp: uint(r.config.PhotometricInterpretation),
		EPSGCode: uint(r.config.EPSGCode)}

	return nil
}

// Retrieve the file name of this GeoTIFF raster file.
func (r *geotiffRaster) FileName() string {
	return r.fileName
}

// Set the file name of this GeoTIFF raster file.
func (r *geotiffRaster) SetFileName(value string) (err error) {
	r.fileName = value
	r.config = NewDefaultRasterConfig()

	// does the file exist?
	if _, err = os.Stat(r.fileName); err == nil {
		// yes it does; read the file
		if err = r.ReadFile(); err != nil {
			return err
		}
	} else {
		return FileDoesNotExistError
	}

	r.minimumValue = math.MaxFloat64
	r.maximumValue = -math.MaxFloat64

	//r.gt = geotiff.GeoTIFF{}
	return nil
}

// Retrieve the RasterType of this Raster.
func (r *geotiffRaster) RasterType() RasterType {
	return RT_GeoTiff
}

// Retrieve the number of rows this raster file.
func (r *geotiffRaster) Rows() int {
	return r.header.rows
}

// Sets the number of rows of this raster file.
func (r *geotiffRaster) SetRows(value int) {
	r.header.rows = value
}

// Retrieve the number of columns of this raster file.
func (r *geotiffRaster) Columns() int {
	return r.header.columns
}

// Sets the number of columns of this raster file.
func (r *geotiffRaster) SetColumns(value int) {
	r.header.columns = value
}

// Retrieve the raster's northern edge's coordinate
func (r *geotiffRaster) North() float64 {
	return r.header.north
}

// Retrieve the raster's southern edge's coordinate
func (r *geotiffRaster) South() float64 {
	return r.header.south
}

// Retrieve the raster's eastern edge's coordinate
func (r *geotiffRaster) East() float64 {
	return r.header.east
}

// Retrieve the raster's western edge's coordinate
func (r *geotiffRaster) West() float64 {
	return r.header.west
}

// Retrieve the raster's minimum value
func (r *geotiffRaster) MinimumValue() float64 {
	if r.minimumValue == math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.minimumValue
}

// Retrieve the raster's minimum value
func (r *geotiffRaster) MaximumValue() float64 {
	if r.maximumValue == -math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.maximumValue
}

func (r *geotiffRaster) findMinAndMaxVals() (minVal float64, maxVal float64) {
	if r.data != nil && len(r.data) > 0 {
		minVal = math.MaxFloat64
		maxVal = -math.MaxFloat64
		for _, v := range r.data {
			if v != r.config.NoDataValue {
				if v > maxVal {
					maxVal = v
				}
				if v < minVal {
					minVal = v
				}
			}
		}
		return minVal, maxVal
	} else {
		return math.MaxFloat64, -math.MaxFloat64
	}
}

// Sets the raster config
func (r *geotiffRaster) SetRasterConfig(value *RasterConfig) {
	r.config = value
}

// Retrieves the raster config
func (r *geotiffRaster) GetRasterConfig() *RasterConfig {
	return r.config
}

// Retrieve the NoData value used by this file.
func (r *geotiffRaster) NoData() float64 {
	return r.config.NoDataValue
}

// Sets the NoData value used by this file.
func (r *geotiffRaster) SetNoData(value float64) {
	r.config.NoDataValue = value
}

// Retrieve the byte order used by this file.
func (r *geotiffRaster) ByteOrder() binary.ByteOrder {
	return r.config.ByteOrder
}

// Sets the byte order used by this file.
func (r *geotiffRaster) SetByteOrder(value binary.ByteOrder) {
	// Do nothing, there is no byte order for ASCII file formats
	// This method is simply present to satisfy the RasterData interface
}

// Retrieves the metadata for this raster
func (r *geotiffRaster) MetadataEntries() []string {
	// This file format does not support metadata. This method
	// is simply present to satisfy the rasterData interface. It will
	// however be used to return the tags for the tiff file.
	ret := make([]string, 1)
	ret[0] = r.gt.GetTags()
	return ret
}

// Adds a metadata entry to this raster
func (r *geotiffRaster) AddMetadataEntry(value string) {
	// This file format does not support metadata. This method
	// is simply present to satisfy the rasterData interface.
}

// Returns the data as a slice of float64 values
func (r *geotiffRaster) Data() ([]float64, error) {
	if len(r.data) == 0 {
		r.ReadFile()
	}
	return r.data, nil
}

// Sets the data from a slice of float64 values
func (r *geotiffRaster) SetData(values []float64) {
	if r.header.numCells == 0 {
		r.header.numCells = r.header.rows * r.header.columns
	}
	if len(values) == r.header.numCells {
		r.data = values
	} else {
		panic(DataSetError)
	}
}

// Returns the value within data
func (r *geotiffRaster) Value(index int) float64 {
	return float64(r.data[index])
}

// Sets the value of index within data
func (r *geotiffRaster) SetValue(index int, value float64) {
	r.data[index] = value
}

// Save the file
func (r *geotiffRaster) Save() (err error) {
	// does the file already exist? If yes, delete it.
	if _, err = os.Stat(r.fileName); err == nil {
		if err = os.Remove(r.fileName); err != nil {
			return FileDeletingError
		}
	}

	r.gt.Data = r.data

	if r.config.PixelIsArea {
		cellSizeX := (r.header.east - r.header.west) / float64(r.header.columns)
		cellSizeY := (r.header.north - r.header.south) / float64(r.header.rows)

		tiepointData := geotiff.TiepointTransformationParameters{I: 0.0, J: 0.0, K: 0.0, X: r.header.west, Y: r.header.north, Z: 0.0, ScaleX: cellSizeX, ScaleY: cellSizeY, ScaleZ: 0.0}
		r.gt.TiepointData = tiepointData
	} else {
		cellSizeX := (r.header.east - r.header.west) / float64(r.header.columns)
		cellSizeY := (r.header.north - r.header.south) / float64(r.header.rows)

		tiepointData := geotiff.TiepointTransformationParameters{I: 0.0, J: 0.0, K: 0.0, X: r.header.west, Y: r.header.north, Z: 0.0, ScaleX: cellSizeX, ScaleY: cellSizeY, ScaleZ: 0.0}
		r.gt.TiepointData = tiepointData
	}

	if r.config.NoDataValue != math.MaxFloat32 {
		r.gt.NodataValue = strconv.FormatFloat(r.config.NoDataValue, 'f', -1, 64)
		r.gt.NodataValue = strings.TrimSpace(r.gt.NodataValue)
		//r.gt.NodataValue = strings.Trim(r.gt.NodataValue, "\x00")

	}

	err = r.gt.Write(r.fileName)
	if err != nil {
		return err
	}
	return nil
}

// Reads the file
func (r *geotiffRaster) ReadFile() error {
	// read the header file
	if r.fileName == "" {
		return FileReadingError
	}

	//r.gt := new(geotiff.GeoTIFF)
	err := r.gt.Read(r.fileName)
	r.check(err)

	r.header.columns = int(r.gt.Columns)
	r.header.rows = int(r.gt.Rows)

	idf, err := r.gt.FindIFDEntryFromName("ModelPixelScaleTag")
	r.check(err)
	modelPixelScale, err := idf.InterpretDataAsFloat()
	r.check(err)

	idf, err = r.gt.FindIFDEntryFromName("ModelTiepointTag")
	r.check(err)
	modelTiepoint, err := idf.InterpretDataAsFloat()
	r.check(err)

	r.header.north = modelTiepoint[4] + modelTiepoint[1]*modelPixelScale[1]
	r.header.south = modelTiepoint[4] - (float64(r.header.rows)-modelTiepoint[1])*modelPixelScale[1]
	r.header.east = modelTiepoint[3] + (float64(r.header.columns)-modelTiepoint[0])*modelPixelScale[0]
	r.header.west = modelTiepoint[3] - modelTiepoint[0]*modelPixelScale[0]

	if r.gt.NodataValue != "" {
		r.config.NoDataValue, err = strconv.ParseFloat(r.gt.NodataValue, 64)
		r.check(err)
	} else {
		r.config.NoDataValue = math.MaxFloat32
	}

	// set the data type based on the sample format and the bitspersample
	numSamples := len(r.gt.BitsPerSample)
	bitDepth := r.gt.BitsPerSample[0]
	sampleFormat := r.gt.SampleFormat
	switch numSamples {
	case 1:
		switch sampleFormat {
		case geotiff.SF_FloatingPoint:
			switch bitDepth {
			case 32:
				r.config.DataType = DT_FLOAT32
			case 64:
				r.config.DataType = DT_FLOAT64
			default:
				panic(errors.New("Unrecognizable data format"))
			}
		case geotiff.SF_UnsignedInteger:
			switch bitDepth {
			case 8:
				r.config.DataType = DT_UINT8
			case 16:
				r.config.DataType = DT_UINT16
			case 32:
				r.config.DataType = DT_UINT32
			case 64:
				r.config.DataType = DT_UINT64
			default:
				panic(errors.New("Unrecognizable data format"))
			}
		case geotiff.SF_SignedInteger:
			switch bitDepth {
			case 8:
				r.config.DataType = DT_INT8
			case 16:
				r.config.DataType = DT_INT16
			case 32:
				r.config.DataType = DT_INT32
			case 64:
				r.config.DataType = DT_INT64
			default:
				panic(errors.New("Unrecognizable data format"))
			}
		default:
			panic(errors.New("Unrecognizable data format"))
		}
	case 3:
		switch bitDepth {
		case 8:
			r.config.DataType = DT_RGB24
		case 16:
			r.config.DataType = DT_RGB48
		default:
			panic(errors.New("Unrecognizable data format"))
		}
	case 4:
		switch bitDepth {
		case 8:
			r.config.DataType = DT_RGBA32
		case 16:
			r.config.DataType = DT_RGBA64
		default:
			panic(errors.New("Unrecognizable data format"))
		}
	default:
		panic(errors.New("Unrecognizable data format"))
	}

	// get the EPSG code of the file
	r.config.EPSGCode = int(r.gt.EPSGCode)

	r.data = r.gt.Data

	return nil
}

type geotiffRasterHeader struct {
	rows     int
	columns  int
	numCells int
	cellSize float64
	north    float64
	south    float64
	east     float64
	west     float64
}

func (r *geotiffRaster) check(e error) {
	if e != nil {
		panic(e)
	}
}
