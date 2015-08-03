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
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Used to manipulate an ArcGIS binary raster (.flt) file.
type arcGisBinaryRaster struct {
	dataFile     string
	data         []float32
	header       arcGisBinaryRasterHeader
	minimumValue float64
	maximumValue float64
	config       *RasterConfig
}

func (r *arcGisBinaryRaster) InitializeRaster(fileName string,
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
	r.header.byteOrder = config.ByteOrder

	// set the file names; if they exist, delete them
	// sort out the names of the header and data files
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == ".flt" {
		r.dataFile = fileName
		r.header.fileName = strings.Replace(fileName, ext, ".hdr", -1)
	} else if ext == ".hdr" {
		r.header.fileName = fileName
		r.dataFile = strings.Replace(fileName, ext, ".flt", -1)
	} else {
		return errors.New("Unrecognized file type.")
	}

	// do the files already exist? If yes, delete them.
	if err = r.deleteFiles(); err != nil {
		return err
	}

	// initialize the data array
	r.data = make([]float32, r.header.numCells)
	if config.InitialValue != 0 {
		initVal := float32(config.InitialValue)
		for i := range r.data {
			r.data[i] = initVal
		}
	}

	r.minimumValue = math.MaxFloat64
	r.maximumValue = -math.MaxFloat64

	return nil
}

// Retrieve the data file name (.flt) of this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) FileName() string {
	return r.dataFile
}

// Set the data file name (.flt) of this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) SetFileName(value string) (err error) {
	r.config = NewDefaultRasterConfig()

	// sort out the names of the header and data files
	ext := strings.ToLower(filepath.Ext(value))
	if ext == ".flt" {
		r.dataFile = value
		r.header.fileName = strings.Replace(value, ext, ".hdr", -1)
	} else if ext == ".hdr" {
		r.header.fileName = value
		r.dataFile = strings.Replace(value, ext, ".flt", -1)
	} else {
		return UnsupportedRasterFormatError
	}

	// does the file exist?
	if _, err = os.Stat(r.header.fileName); err == nil {
		// yes it does; read the file
		if err = r.ReadFile(); err != nil {
			return err
		}
	} else {
		return FileDoesNotExistError
	}

	r.minimumValue = math.MaxFloat64
	r.maximumValue = -math.MaxFloat64

	return nil
}

// Retrieve the RasterType of this Raster.
func (r *arcGisBinaryRaster) RasterType() RasterType {
	return RT_ArcGisBinaryRaster
}

// Retrieve the number of rows this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) Rows() int {
	return r.header.rows
}

// Sets the number of rows of this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) SetRows(value int) {
	r.header.rows = value
}

// Retrieve the number of columns of this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) Columns() int {
	return r.header.columns
}

// Sets the number of columns of this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) SetColumns(value int) {
	r.header.columns = value
}

// Retrieve the raster's northern edge's coordinate
func (r *arcGisBinaryRaster) North() float64 {
	return r.header.north
}

// Retrieve the raster's southern edge's coordinate
func (r *arcGisBinaryRaster) South() float64 {
	return r.header.south
}

// Retrieve the raster's eastern edge's coordinate
func (r *arcGisBinaryRaster) East() float64 {
	return r.header.east
}

// Retrieve the raster's western edge's coordinate
func (r *arcGisBinaryRaster) West() float64 {
	return r.header.west
}

// Retrieve the raster's minimum value
func (r *arcGisBinaryRaster) MinimumValue() float64 {
	if r.minimumValue == math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.minimumValue
}

// Retrieve the raster's minimum value
func (r *arcGisBinaryRaster) MaximumValue() float64 {
	if r.maximumValue == -math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.maximumValue
}

func (r *arcGisBinaryRaster) findMinAndMaxVals() (minVal float64, maxVal float64) {
	if r.data != nil && len(r.data) > 0 {
		minVal = math.MaxFloat64
		maxVal = -math.MaxFloat64
		for _, v := range r.data {
			v2 := float64(v)
			if v2 != r.header.nodata {
				if v2 > maxVal {
					maxVal = v2
				}
				if v2 < minVal {
					minVal = v2
				}
			}
		}
		return minVal, maxVal
	} else {
		return math.MaxFloat64, -math.MaxFloat64
	}
}

// Sets the raster config
func (r *arcGisBinaryRaster) SetRasterConfig(value *RasterConfig) {
	r.config = value
}

// Retrieves the raster config
func (r *arcGisBinaryRaster) GetRasterConfig() *RasterConfig {
	return r.config
}

// Retrieve the NoData value used by this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) NoData() float64 {
	return r.header.nodata
}

// Sets the NoData value used by this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) SetNoData(value float64) {
	r.header.nodata = value
}

// Retrieve the byte order used by this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) ByteOrder() binary.ByteOrder {
	return r.header.byteOrder
}

// Sets the byte order used by this ArcGIS binary raster file.
func (r *arcGisBinaryRaster) SetByteOrder(value binary.ByteOrder) {
	r.header.byteOrder = value
}

// Retrieves the metadata for this raster
func (r *arcGisBinaryRaster) MetadataEntries() []string {
	// This file format does not support metadata. This method
	// is simply present to satisfy the RasterData interface.
	return nil
}

// Adds a metadata entry to this raster
func (r *arcGisBinaryRaster) AddMetadataEntry(value string) {
	// This file format does not support metadata. This method
	// is simply present to satisfy the RasterData interface.
}

// Returns the data as a slice of float64 values
func (r *arcGisBinaryRaster) Data() ([]float64, error) {
	if len(r.data) == 0 {
		r.ReadFile()
	}
	// convert the float32 to a float64
	retData := make([]float64, r.header.numCells)
	for i, v := range r.data {
		retData[i] = float64(v)
	}
	return retData, nil
}

// Sets the data from a slice of float64 values
func (r *arcGisBinaryRaster) SetData(values []float64) {
	// make sure that the numCells is set
	if r.header.numCells == 0 {
		r.header.numCells = r.header.rows * r.header.columns
	}
	if len(values) == r.header.numCells {
		// convert the float32 to a float64
		r.data = make([]float32, r.header.numCells)
		for i, v := range values {
			r.data[i] = float32(v)
		}
	} else {
		panic(DataSetError)
	}
}

// Returns the value within data
func (r *arcGisBinaryRaster) Value(index int) float64 {
	return float64(r.data[index])
}

// Sets the value of index within data
func (r *arcGisBinaryRaster) SetValue(index int, value float64) {
	r.data[index] = float32(value)
}

//// Returns the value within ColorData
//func (r *arcGisBinaryRaster) GetColor(index int) color.Color {
//	// Return black, this raster format does not support RGB colour.
//	return color.RGBA{0, 0, 0, 0}
//}

//// Sets the value of index within ColorData
//func (r *arcGisBinaryRaster) SetColor(index int, value color.Color) {
//	// do nothing, this raster format does not support RGB colour.
//}

// Save the file
func (r *arcGisBinaryRaster) Save() (err error) {
	// do the files exist? If yes, delete them.
	if err = r.deleteFiles(); err != nil {
		return err
	}

	// write the header file
	if err = r.header.writeHeaderFile(); err != nil {
		return err
	}

	// write the data file
	f, err := os.Create(r.dataFile)
	r.check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	//buf := new(bytes.Buffer)
	for _, v := range r.data {
		if err = binary.Write(w, r.header.byteOrder, v); err != nil {
			return err
		}
	}
	//w.Write(buf.Bytes())
	w.Flush()
	return nil
}

// Reads the file
func (r *arcGisBinaryRaster) ReadFile() error {
	// read the header file
	err := r.header.readHeaderFile()
	if err != nil {
		return errors.New("ArcGIS binary raster header file not read properly.")
	}

	// read the data file
	bytedata, err := ioutil.ReadFile(r.dataFile)
	buf := bytes.NewReader(bytedata)
	r.header.numCells = r.header.columns * r.header.rows
	r.data = make([]float32, r.header.numCells)
	err = binary.Read(buf, r.header.byteOrder, &r.data)
	r.check(err)

	return nil
}

type arcGisBinaryRasterHeader struct {
	fileName       string
	rows           int
	columns        int
	numCells       int
	nodata         float64
	cellSize       float64
	north          float64
	south          float64
	east           float64
	west           float64
	byteOrder      binary.ByteOrder
	cellCornerMode bool
}

func (h *arcGisBinaryRasterHeader) readHeaderFile() error {
	// read the header file
	if h.fileName == "" {
		return errors.New("ArcGIS binary raster header file not set properly.")
	}
	content, err := ioutil.ReadFile(h.fileName)
	h.check(err)
	str := strings.Replace(string(content), "\r\n", "\n", -1)
	var xllcenter float64
	var yllcenter float64
	var xllcorner float64
	var yllcorner float64
	lines := strings.Split(str, "\n")
	for a := 0; a < len(lines); a++ {
		//fmt.Println(lines[a])
		str = strings.ToLower(lines[a])
		if strings.Contains(str, "ncols") {
			s := strings.Fields(str)
			h.columns, err = strconv.Atoi(s[len(s)-1])
			h.check(err)
		} else if strings.Contains(str, "nrows") {
			s := strings.Fields(str)
			h.rows, err = strconv.Atoi(s[len(s)-1])
			h.check(err)
		} else if strings.Contains(str, "byteorder") {
			s := strings.Fields(str)
			if strings.Contains(s[len(s)-1], "lsb") {
				h.byteOrder = binary.LittleEndian
			} else {
				h.byteOrder = binary.BigEndian
			}
		} else if strings.Contains(str, "nodata") {
			s := strings.Fields(str)
			h.nodata, err = strconv.ParseFloat(s[len(s)-1], 64)
			h.check(err)
		} else if strings.Contains(str, "cellsize") {
			s := strings.Fields(str)
			h.cellSize, err = strconv.ParseFloat(s[len(s)-1], 64)
			h.check(err)
		} else if strings.Contains(str, "xllcenter") {
			s := strings.Fields(str)
			xllcenter, err = strconv.ParseFloat(s[len(s)-1], 64)
			h.check(err)
		} else if strings.Contains(str, "yllcenter") {
			s := strings.Fields(str)
			yllcenter, err = strconv.ParseFloat(s[len(s)-1], 64)
			h.check(err)
		} else if strings.Contains(str, "xllcorner") {
			s := strings.Fields(str)
			xllcorner, err = strconv.ParseFloat(s[len(s)-1], 64)
			h.check(err)
		} else if strings.Contains(str, "yllcorner") {
			s := strings.Fields(str)
			yllcorner, err = strconv.ParseFloat(s[len(s)-1], 64)
			h.check(err)
		}
	}
	//set the North, East, South, and West coodinates
	if xllcorner != 0 {
		h.cellCornerMode = true
		h.east = xllcorner + float64(h.columns)*h.cellSize
		h.west = xllcorner
		h.south = yllcorner
		h.north = yllcorner + float64(h.rows)*h.cellSize
	} else {
		h.cellCornerMode = false
		h.east = xllcenter - (0.5 * h.cellSize) + float64(h.columns)*h.cellSize
		h.west = xllcenter - (0.5 * h.cellSize)
		h.south = yllcenter - (0.5 * h.cellSize)
		h.north = yllcenter - (0.5 * h.cellSize) + float64(h.rows)*h.cellSize
	}
	return nil
}

func (h *arcGisBinaryRasterHeader) writeHeaderFile() (err error) {
	f, err := os.Create(h.fileName)
	h.check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	var str string
	str = "NCOLS         " + strconv.Itoa(h.columns)
	_, err = w.WriteString(str + "\n")
	h.check(err)
	str = "NROWS         " + strconv.Itoa(h.rows)
	_, err = w.WriteString(str + "\n")
	h.check(err)
	if h.cellCornerMode {
		str = "XLLCORNER     " + strconv.FormatFloat(h.west, 'f', -1, 64)
		_, err = w.WriteString(str + "\n")
		h.check(err)
		str = "YLLCORNER     " + strconv.FormatFloat(h.south, 'f', -1, 64)
		_, err = w.WriteString(str + "\n")
		h.check(err)
	} else {
		str = "XLLCENTER     " + strconv.FormatFloat(h.west+h.cellSize/2.0, 'f', -1, 64)
		_, err = w.WriteString(str + "\n")
		h.check(err)
		str = "YLLCENTER     " + strconv.FormatFloat(h.south+h.cellSize/2.0, 'f', -1, 64)
		_, err = w.WriteString(str + "\n")
		h.check(err)
	}
	str = "CELLSIZE      " + strconv.FormatFloat(h.cellSize, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	h.check(err)
	str = "NODATA_VALUE  " + strconv.FormatFloat(h.nodata, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	h.check(err)
	if h.byteOrder == binary.LittleEndian {
		str = "BYTEORDER     lsbfirst"
		_, err = w.WriteString(str + "\n")
		h.check(err)
	} else {
		str = "BYTEORDER     msbfirst"
		_, err = w.WriteString(str + "\n")
		h.check(err)
	}
	w.Flush()
	return nil
}

func (r *arcGisBinaryRaster) check(e error) {
	if e != nil {
		panic(e)
	}
}

func (h *arcGisBinaryRasterHeader) check(e error) {
	if e != nil {
		panic(e)
	}
}

func (r *arcGisBinaryRaster) deleteFiles() (err error) {
	// do the files exist?
	if _, err = os.Stat(r.header.fileName); err == nil {
		if err = os.Remove(r.header.fileName); err != nil {
			return FileDeletingError
		}
	}
	if _, err = os.Stat(r.dataFile); err == nil {
		if err = os.Remove(r.dataFile); err != nil {
			return FileDeletingError
		}
	}
	return nil
}
