// Copyright 2014 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Nov. 2014.

// Package raster provides support for reading and creating various common
// geospatial raster data formats.
package raster

import (
	"bufio"
	"encoding/binary"
	"math"
	"os"
	"strconv"
	"strings"
)

// Used to manipulate an ArcGIS ASCII raster file.
type arcGisAsciiRaster struct {
	fileName     string
	data         []float64
	header       arcGisAsciiRasterHeader
	minimumValue float64
	maximumValue float64
	config       *RasterConfig
}

func (r *arcGisAsciiRaster) InitializeRaster(fileName string,
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
	r.header.cellCornerMode = true
	r.header.cellSize = (east - west) / float64(r.header.columns)
	r.header.nodata = config.NoDataValue

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

	return nil
}

// Retrieve the file name of this ArcGIS ASCII raster file.
func (r *arcGisAsciiRaster) FileName() string {
	return r.fileName
}

// Set the file name of this ArcGIS ASCII raster file.
func (r *arcGisAsciiRaster) SetFileName(value string) (err error) {
	r.config = NewDefaultRasterConfig()

	r.fileName = value
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
	r.config.RasterFormat = RT_ArcGisAsciiRaster

	return nil
}

// Retrieve the RasterType of this Raster.
func (r *arcGisAsciiRaster) RasterType() RasterType {
	return RT_ArcGisAsciiRaster
}

// Retrieve the number of rows this ArcGIS binary raster file.
func (r *arcGisAsciiRaster) Rows() int {
	return r.header.rows
}

// Sets the number of rows of this ArcGIS binary raster file.
func (r *arcGisAsciiRaster) SetRows(value int) {
	r.header.rows = value
}

// Retrieve the number of columns of this ArcGIS binary raster file.
func (r *arcGisAsciiRaster) Columns() int {
	return r.header.columns
}

// Sets the number of columns of this ArcGIS binary raster file.
func (r *arcGisAsciiRaster) SetColumns(value int) {
	r.header.columns = value
}

// Retrieve the raster's northern edge's coordinate
func (r *arcGisAsciiRaster) North() float64 {
	return r.header.north
}

// Retrieve the raster's southern edge's coordinate
func (r *arcGisAsciiRaster) South() float64 {
	return r.header.south
}

// Retrieve the raster's eastern edge's coordinate
func (r *arcGisAsciiRaster) East() float64 {
	return r.header.east
}

// Retrieve the raster's western edge's coordinate
func (r *arcGisAsciiRaster) West() float64 {
	return r.header.west
}

// Retrieve the raster's minimum value
func (r *arcGisAsciiRaster) MinimumValue() float64 {
	if r.minimumValue == math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.minimumValue
}

// Retrieve the raster's minimum value
func (r *arcGisAsciiRaster) MaximumValue() float64 {
	if r.maximumValue == -math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.maximumValue
}

func (r *arcGisAsciiRaster) findMinAndMaxVals() (minVal float64, maxVal float64) {
	if r.data != nil && len(r.data) > 0 {
		minVal = math.MaxFloat64
		maxVal = -math.MaxFloat64
		for _, v := range r.data {
			if v != r.header.nodata {
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
func (r *arcGisAsciiRaster) SetRasterConfig(value *RasterConfig) {
	r.config = value
}

// Retrieves the raster config
func (r *arcGisAsciiRaster) GetRasterConfig() *RasterConfig {
	return r.config
}

// Retrieve the NoData value used by this ArcGIS binary raster file.
func (r *arcGisAsciiRaster) NoData() float64 {
	return r.header.nodata
}

// Sets the NoData value used by this ArcGIS binary raster file.
func (r *arcGisAsciiRaster) SetNoData(value float64) {
	r.header.nodata = value
}

// Retrieve the byte order used by this ArcGIS binary raster file.
func (r *arcGisAsciiRaster) ByteOrder() binary.ByteOrder {
	return nil
}

// Sets the byte order used by this ArcGIS binary raster file.
func (r *arcGisAsciiRaster) SetByteOrder(value binary.ByteOrder) {
	// Do nothing, there is no byte order for ASCII file formats
	// This method is simply present to satisfy the RasterData interface
}

// Retrieves the metadata for this raster
func (r *arcGisAsciiRaster) MetadataEntries() []string {
	// This file format does not support metadata. This method
	// is simply present to satisfy the rasterData interface.
	return nil
}

// Adds a metadata entry to this raster
func (r *arcGisAsciiRaster) AddMetadataEntry(value string) {
	// This file format does not support metadata. This method
	// is simply present to satisfy the rasterData interface.
}

// Returns the data as a slice of float64 values
func (r *arcGisAsciiRaster) Data() ([]float64, error) {
	if len(r.data) == 0 {
		r.ReadFile()
	}
	return r.data, nil
}

// Sets the data from a slice of float64 values
func (r *arcGisAsciiRaster) SetData(values []float64) {
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
func (r *arcGisAsciiRaster) Value(index int) float64 {
	return float64(r.data[index])
}

// Sets the value of index within data
func (r *arcGisAsciiRaster) SetValue(index int, value float64) {
	r.data[index] = value
}

//// Returns the value within ColorData
//func (r *arcGisAsciiRaster) GetColor(index int) color.Color {
//	// Return black, this raster format does not support RGB colour.
//	return color.RGBA{0, 0, 0, 0}
//}

//// Sets the value of index within ColorData
//func (r *arcGisAsciiRaster) SetColor(index int, value color.Color) {
//	// do nothing, this raster format does not support RGB colour.
//}

// Save the file
func (r *arcGisAsciiRaster) Save() (err error) {
	// does the file already exist? If yes, delete it.
	if _, err = os.Stat(r.fileName); err == nil {
		if err = os.Remove(r.fileName); err != nil {
			return FileDeletingError
		}
	}

	// write the header file
	f, err := os.Create(r.fileName)
	r.check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	var str string
	str = "NCOLS         " + strconv.Itoa(r.header.columns)
	_, err = w.WriteString(str + "\n")
	r.check(err)
	str = "NROWS         " + strconv.Itoa(r.header.rows)
	w.WriteString(str + "\n")
	if r.header.cellCornerMode {
		str = "XLLCORNER     " + strconv.FormatFloat(r.header.west, 'f', -1, 64)
		w.WriteString(str + "\n")
		str = "YLLCORNER     " + strconv.FormatFloat(r.header.south, 'f', -1, 64)
		w.WriteString(str + "\n")
	} else {
		str = "XLLCENTER     " + strconv.FormatFloat(r.header.west+r.header.cellSize/2.0, 'f', -1, 64)
		w.WriteString(str + "\n")
		str = "YLLCENTER     " + strconv.FormatFloat(r.header.south+r.header.cellSize/2.0, 'f', -1, 64)
		w.WriteString(str + "\n")
	}
	str = "CELLSIZE      " + strconv.FormatFloat(r.header.cellSize, 'f', -1, 64)
	w.WriteString(str + "\n")
	str = "NODATA_VALUE  " + strconv.FormatFloat(r.header.nodata, 'f', -1, 64)
	w.WriteString(str + "\n")
	cellNum := 0
	for row := 0; row < r.header.rows; row++ {
		str = ""
		for col := 0; col < r.header.columns; col++ {
			str += strconv.FormatFloat(r.data[cellNum], 'f', -1, 64) + " "
			cellNum++
		}
		str = strings.TrimSpace(str) + "\n"
		w.WriteString(str)
	}

	w.Flush()

	// write the data file

	return nil
}

// Reads the file
func (r *arcGisAsciiRaster) ReadFile() error {
	// read the header file
	if r.fileName == "" {
		return FileReadingError
	}

	// read in the first six lines of the file
	var xllcenter float64
	var yllcenter float64
	var xllcorner float64
	var yllcorner float64

	f, err := os.Open(r.fileName)
	if err != nil {
		return FileOpeningError
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	cellNum := 0
	for scanner.Scan() {
		str := strings.ToLower(scanner.Text())
		lineNum++
		if lineNum <= 6 {
			if strings.Contains(str, "ncols") {
				s := strings.Fields(str)
				r.header.columns, err = strconv.Atoi(s[len(s)-1])
				r.check(err)
				if r.header.rows > 0 {
					r.header.numCells = r.header.columns * r.header.rows
					r.data = make([]float64, r.header.numCells)
				}
			} else if strings.Contains(str, "nrows") {
				s := strings.Fields(str)
				r.header.rows, err = strconv.Atoi(s[len(s)-1])
				r.check(err)
				if r.header.columns > 0 {
					r.header.numCells = r.header.columns * r.header.rows
					r.data = make([]float64, r.header.numCells)
				}
			} else if strings.Contains(str, "nodata") {
				s := strings.Fields(str)
				r.header.nodata, err = strconv.ParseFloat(s[len(s)-1], 64)
				r.check(err)
			} else if strings.Contains(str, "cellsize") {
				s := strings.Fields(str)
				r.header.cellSize, err = strconv.ParseFloat(s[len(s)-1], 64)
				r.check(err)
			} else if strings.Contains(str, "xllcenter") {
				s := strings.Fields(str)
				xllcenter, err = strconv.ParseFloat(s[len(s)-1], 64)
				r.check(err)
			} else if strings.Contains(str, "yllcenter") {
				s := strings.Fields(str)
				yllcenter, err = strconv.ParseFloat(s[len(s)-1], 64)
				r.check(err)
			} else if strings.Contains(str, "xllcorner") {
				s := strings.Fields(str)
				xllcorner, err = strconv.ParseFloat(s[len(s)-1], 64)
				r.check(err)
			} else if strings.Contains(str, "yllcorner") {
				s := strings.Fields(str)
				yllcorner, err = strconv.ParseFloat(s[len(s)-1], 64)
				r.check(err)
			}
		} else { // it's a data line
			s := strings.Fields(str)
			for _, v := range s {
				r.data[cellNum], _ = strconv.ParseFloat(v, 64)
				cellNum++
			}
		}
	}

	//set the North, East, South, and West coodinates
	if xllcorner != 0 {
		r.header.cellCornerMode = true
		r.header.east = xllcorner + float64(r.header.columns)*r.header.cellSize
		r.header.west = xllcorner
		r.header.south = yllcorner
		r.header.north = yllcorner + float64(r.header.rows)*r.header.cellSize
	} else {
		r.header.cellCornerMode = false
		r.header.east = xllcenter - (0.5 * r.header.cellSize) + float64(r.header.columns)*r.header.cellSize
		r.header.west = xllcenter - (0.5 * r.header.cellSize)
		r.header.south = yllcenter - (0.5 * r.header.cellSize)
		r.header.north = yllcenter - (0.5 * r.header.cellSize) + float64(r.header.rows)*r.header.cellSize
	}

	return nil
}

type arcGisAsciiRasterHeader struct {
	rows           int
	columns        int
	numCells       int
	nodata         float64
	cellSize       float64
	north          float64
	south          float64
	east           float64
	west           float64
	cellCornerMode bool
}

func (r *arcGisAsciiRaster) check(e error) {
	if e != nil {
		panic(e)
	}
}
