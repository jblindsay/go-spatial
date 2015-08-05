// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Aug. 2015.

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

// Used to manipulate an Idrisi raster (.rst) file.
type idrisiRaster struct {
	dataFile     string
	data         []float64
	header       idrisiRasterHeader
	minimumValue float64
	maximumValue float64
	config       *RasterConfig
}

func (r *idrisiRaster) InitializeRaster(fileName string,
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
	r.header.nodata = config.NoDataValue
	r.config.ByteOrder = config.ByteOrder
	r.config.RasterFormat = RT_IdrisiRaster

	// set the file names; if they exist, delete them
	// sort out the names of the header and data files
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == ".rst" {
		r.dataFile = fileName
		r.header.fileName = strings.Replace(fileName, ext, ".rdc", -1)
	} else if ext == ".rdc" {
		r.header.fileName = fileName
		r.dataFile = strings.Replace(fileName, ext, ".rst", -1)
	} else {
		return errors.New("Unrecognized file type.")
	}

	// do the files already exist? If yes, delete them.
	if err = r.deleteFiles(); err != nil {
		return err
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

// Retrieve the data file name (.rst) of this Idrisi raster file.
func (r *idrisiRaster) FileName() string {
	return r.dataFile
}

// Set the data file name (.rst) of this Idrisi raster file.
func (r *idrisiRaster) SetFileName(value string) (err error) {
	r.config = NewDefaultRasterConfig()

	// sort out the names of the header and data files
	ext := strings.ToLower(filepath.Ext(value))
	if ext == ".rst" {
		r.dataFile = value
		r.header.fileName = strings.Replace(value, ext, ".rdc", -1)
	} else if ext == ".rdc" {
		r.header.fileName = value
		r.dataFile = strings.Replace(value, ext, ".rst", -1)
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
	r.config.RasterFormat = RT_IdrisiRaster

	return nil
}

// Retrieve the RasterType of this Raster.
func (r *idrisiRaster) RasterType() RasterType {
	return RT_IdrisiRaster
}

// Retrieve the number of rows this binary raster file.
func (r *idrisiRaster) Rows() int {
	return r.header.rows
}

// Sets the number of rows of this binary raster file.
func (r *idrisiRaster) SetRows(value int) {
	r.header.rows = value
}

// Retrieve the number of columns of this binary raster file.
func (r *idrisiRaster) Columns() int {
	return r.header.columns
}

// Sets the number of columns of this binary raster file.
func (r *idrisiRaster) SetColumns(value int) {
	r.header.columns = value
}

// Retrieve the raster's northern edge's coordinate
func (r *idrisiRaster) North() float64 {
	return r.header.north
}

// Retrieve the raster's southern edge's coordinate
func (r *idrisiRaster) South() float64 {
	return r.header.south
}

// Retrieve the raster's eastern edge's coordinate
func (r *idrisiRaster) East() float64 {
	return r.header.east
}

// Retrieve the raster's western edge's coordinate
func (r *idrisiRaster) West() float64 {
	return r.header.west
}

// Retrieve the raster's minimum value
func (r *idrisiRaster) MinimumValue() float64 {
	if r.minimumValue == math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.minimumValue
}

// Retrieve the raster's minimum value
func (r *idrisiRaster) MaximumValue() float64 {
	if r.maximumValue == -math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.maximumValue
}

func (r *idrisiRaster) findMinAndMaxVals() (minVal float64, maxVal float64) {
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
func (r *idrisiRaster) SetRasterConfig(value *RasterConfig) {
	r.config = value
}

// Retrieves the raster config
func (r *idrisiRaster) GetRasterConfig() *RasterConfig {
	return r.config
}

// Retrieve the NoData value used by this binary raster file.
func (r *idrisiRaster) NoData() float64 {
	return r.header.nodata
}

// Sets the NoData value used by this binary raster file.
func (r *idrisiRaster) SetNoData(value float64) {
	r.header.nodata = value
	r.config.NoDataValue = value
}

// Retrieve the byte order used by this binary raster file.
func (r *idrisiRaster) ByteOrder() binary.ByteOrder {
	return r.config.ByteOrder
}

// Sets the byte order used by this binary raster file.
func (r *idrisiRaster) SetByteOrder(value binary.ByteOrder) {
	r.config.ByteOrder = value
}

// Retrieves the metadata for this raster
func (r *idrisiRaster) MetadataEntries() []string {
	return r.config.MetadataEntries
}

// Adds a metadata entry to this raster
func (r *idrisiRaster) AddMetadataEntry(value string) {
	mde := r.config.MetadataEntries
	newSlice := make([]string, len(mde)+1)
	for i, val := range mde {
		if len(strings.TrimSpace(val)) > 0 {
			newSlice[i] = val
		}
	}
	newSlice[len(mde)] = value
	r.config.MetadataEntries = newSlice
}

// Returns the data as a slice of float64 values
func (r *idrisiRaster) Data() ([]float64, error) {
	if len(r.data) == 0 {
		r.ReadFile()
	}
	return r.data, nil
}

// Sets the data from a slice of float64 values
func (r *idrisiRaster) SetData(values []float64) {
	// make sure that the numCells is set
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
func (r *idrisiRaster) Value(index int) float64 {
	return r.data[index]
}

// Sets the value of index within data
func (r *idrisiRaster) SetValue(index int, value float64) {
	r.data[index] = value
}

// Save the file
func (r *idrisiRaster) Save() (err error) {
	// do the files exist? If yes, delete them.
	if err = r.deleteFiles(); err != nil {
		return err
	}

	// write the header file
	if err = r.writeHeaderFile(); err != nil {
		return err
	}

	// write the data file
	f, err := os.Create(r.dataFile)
	r.check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	buf := new(bytes.Buffer)
	switch r.config.DataType {
	case DT_FLOAT32:
		out := make([]float32, len(r.data))
		for i := 0; i < len(r.data); i++ {
			out[i] = float32(r.data[i])
		}
		if err = binary.Write(buf, r.config.ByteOrder, out); err != nil {
			return FileWritingError
		}
	case DT_INT16:
		out := make([]int16, len(r.data))
		for i := 0; i < len(r.data); i++ {
			out[i] = int16(r.data[i])
		}
		if err = binary.Write(buf, r.config.ByteOrder, out); err != nil {
			return FileWritingError
		}
	case DT_UINT8:
		out := make([]uint8, len(r.data))
		for i := 0; i < len(r.data); i++ {
			out[i] = uint8(r.data[i])
		}
		if err = binary.Write(buf, r.config.ByteOrder, out); err != nil {
			return FileWritingError
		}
	case DT_RGB24:
		panic("RGB24 data format is not supported")
	default:
		return FileWritingError
	}
	w.Write(buf.Bytes())
	w.Flush()
	return nil
}

// Reads the file
func (r *idrisiRaster) ReadFile() error {
	// read the header file
	err := r.readHeaderFile()
	if err != nil {
		return FileReadingError
	}

	// read the data file
	bytedata, err := ioutil.ReadFile(r.dataFile)
	buf := bytes.NewReader(bytedata)
	r.header.numCells = r.header.columns * r.header.rows
	r.data = make([]float64, r.header.numCells)
	switch r.config.DataType {
	case DT_FLOAT32:
		nativeData := make([]float32, r.header.numCells)
		err = binary.Read(buf, r.config.ByteOrder, &nativeData)
		r.check(err)
		for i, value := range nativeData {
			r.data[i] = float64(value)
		}
		nativeData = nil
	case DT_INT16:
		nativeData := make([]int16, r.header.numCells)
		err = binary.Read(buf, r.config.ByteOrder, &nativeData)
		r.check(err)
		for i, value := range nativeData {
			r.data[i] = float64(value)
		}
		nativeData = nil
	case DT_UINT8:
		nativeData := make([]uint8, r.header.numCells)
		err = binary.Read(buf, r.config.ByteOrder, &nativeData)
		r.check(err)
		for i, value := range nativeData {
			r.data[i] = float64(value)
		}
		nativeData = nil
	case DT_RGB24:
		panic("The RGB24 data type is not currently supported.")
	default:
		return FileReadingError
	}

	return nil
}

type idrisiRasterHeader struct {
	fileName string
	rows     int
	columns  int
	numCells int
	nodata   float64
	north    float64
	south    float64
	east     float64
	west     float64
}

func (r *idrisiRaster) readHeaderFile() error {
	r.header.nodata = -math.MaxFloat64
	// read the header file
	if r.header.fileName == "" {
		return errors.New("Idrisi raster header file not set properly.")
	}
	content, err := ioutil.ReadFile(r.header.fileName)
	r.check(err)
	str := strings.Replace(string(content), "\r\n", "\n", -1)
	lines := strings.Split(str, "\n")
	for a := 0; a < len(lines); a++ {
		str = strings.ToLower(lines[a])
		//println(str)
		s := strings.Split(lines[a], ":")
		if strings.Contains(str, "min. value") && !strings.Contains(str, "lineage") {
			r.minimumValue, err = strconv.ParseFloat(strings.TrimSpace(s[len(s)-1]), 64)
			r.check(err)
		} else if strings.Contains(str, "max. value") && !strings.Contains(str, "lineage") {
			r.maximumValue, err = strconv.ParseFloat(strings.TrimSpace(s[len(s)-1]), 64)
			r.check(err)
		} else if strings.Contains(str, "display min") && !strings.Contains(str, "lineage") {
			r.config.DisplayMinimum, err = strconv.ParseFloat(strings.TrimSpace(s[len(s)-1]), 64)
			r.check(err)
		} else if strings.Contains(str, "display max") && !strings.Contains(str, "lineage") {
			r.config.DisplayMaximum, err = strconv.ParseFloat(strings.TrimSpace(s[len(s)-1]), 64)
			r.check(err)
		} else if strings.Contains(str, "max. y") && !strings.Contains(str, "lineage") {
			r.header.north, err = strconv.ParseFloat(strings.TrimSpace(s[len(s)-1]), 64)
			r.check(err)
		} else if strings.Contains(str, "min. y") && !strings.Contains(str, "lineage") {
			r.header.south, err = strconv.ParseFloat(strings.TrimSpace(s[len(s)-1]), 64)
			r.check(err)
		} else if strings.Contains(str, "max. x") && !strings.Contains(str, "lineage") {
			r.header.east, err = strconv.ParseFloat(strings.TrimSpace(s[len(s)-1]), 64)
			r.check(err)
		} else if strings.Contains(str, "min. x") && !strings.Contains(str, "lineage") {
			r.header.west, err = strconv.ParseFloat(strings.TrimSpace(s[len(s)-1]), 64)
			r.check(err)
		} else if strings.Contains(str, "columns") && !strings.Contains(str, "lineage") {
			r.header.columns, err = strconv.Atoi(strings.TrimSpace(s[len(s)-1]))
			r.check(err)
		} else if strings.Contains(str, "rows") && !strings.Contains(str, "lineage") {
			r.header.rows, err = strconv.Atoi(strings.TrimSpace(s[len(s)-1]))
			r.check(err)
		} else if strings.Contains(str, "data type") && !strings.Contains(str, "lineage") {
			dt := strings.ToLower(strings.TrimSpace(s[len(s)-1]))
			if strings.Contains(dt, "real") {
				r.config.DataType = DT_FLOAT32
			} else if strings.Contains(dt, "int") {
				r.config.DataType = DT_INT16
			} else if strings.Contains(dt, "byte") {
				r.config.DataType = DT_UINT8
			} else if strings.Contains(dt, "rgb24") {
				r.config.DataType = DT_RGB24
			}
		} else if strings.Contains(str, "value units") && !strings.Contains(str, "lineage") {
			r.config.ZUnits = strings.ToLower(strings.TrimSpace(s[len(s)-1]))
		} else if strings.Contains(str, "ref.") && strings.Contains(str, "units") && !strings.Contains(str, "lineage") {
			r.config.XYUnits = strings.ToLower(strings.TrimSpace(s[len(s)-1]))
		} else if strings.Contains(str, "ref.") && strings.Contains(str, "system") && !strings.Contains(str, "lineage") {
			r.config.CoordinateRefSystemWKT = strings.TrimSpace(s[len(s)-1])
		} else if strings.Contains(str, "byteorder") && !strings.Contains(str, "lineage") {
			if strings.Contains(strings.ToLower(s[len(s)-1]), "LITTLE_ENDIAN") {
				r.config.ByteOrder = binary.LittleEndian
			} else {
				r.config.ByteOrder = binary.BigEndian
			}
		} else if strings.Contains(str, "lineage") || strings.Contains(str, "comment") {
			value := strings.TrimSpace(s[len(s)-1])
			value = strings.Replace(value, ";", ":", -1)
			r.AddMetadataEntry(value)
			//r.config.MetadataEntries = append(r.config.MetadataEntries, value)
		} else if strings.Contains(str, "file type") && !strings.Contains(str, "lineage") {
			if !strings.Contains(s[len(s)-1], "binary") || strings.Contains(s[len(s)-1], "packed") {
				panic("Idrisi ASCII and packed binary files are currently unsupported.")
			}
		}
	}

	r.header.numCells = r.header.rows * r.header.columns

	return nil
}

func (r *idrisiRaster) writeHeaderFile() (err error) {
	f, err := os.Create(r.header.fileName)
	r.check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	var str string

	r.minimumValue, r.maximumValue = r.findMinAndMaxVals()

	str = "file format : IDRISI Raster A.1"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "file title  : "
	_, err = w.WriteString(str + "\n")
	r.check(err)

	switch r.config.DataType {
	case DT_FLOAT32:
		str = "data type   : real"
	case DT_INT16:
		str = "data type   : integer"
	case DT_UINT8:
		str = "data type   : byte"
	case DT_RGB24:
		str = "data type   : RGB24"
	}
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "file type   : binary"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "columns     : " + strconv.Itoa(r.header.columns)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "rows        : " + strconv.Itoa(r.header.rows)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "ref. system : " + r.config.CoordinateRefSystemWKT
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "ref. units  : " + r.config.XYUnits
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "unit dist.  : 1.0000000"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "min. X      : " + strconv.FormatFloat(r.header.west, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "max. X      : " + strconv.FormatFloat(r.header.east, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "min. Y      : " + strconv.FormatFloat(r.header.south, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "max. Y      : " + strconv.FormatFloat(r.header.north, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "pos'n error : unknown"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "resolution  : unknown"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "min. value  : " + strconv.FormatFloat(r.minimumValue, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "max. value  : " + strconv.FormatFloat(r.maximumValue, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	if r.config.DisplayMinimum == math.MaxFloat64 {
		r.config.DisplayMinimum = r.minimumValue
	}
	str = "display min : " + strconv.FormatFloat(r.config.DisplayMinimum, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	if r.config.DisplayMaximum == -math.MaxFloat64 {
		r.config.DisplayMaximum = r.maximumValue
	}
	str = "display max : " + strconv.FormatFloat(r.config.DisplayMaximum, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "value units : " + r.config.ZUnits
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "value error : unknown"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "flag value  : " + "none"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "flag def'n  : " + "none"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "legend cats : 0"
	_, err = w.WriteString(str + "\n")
	r.check(err)

	// write the metadata entries
	for _, value := range r.config.MetadataEntries {
		if len(strings.TrimSpace(value)) > 0 {
			str = "comment     : " + strings.Replace(value, ":", ";", -1)
			_, err = w.WriteString(str + "\n")
			r.check(err)
		}
	}

	w.Flush()
	return nil
}

func (r *idrisiRaster) check(e error) {
	if e != nil {
		panic(e)
	}
}

func (h *idrisiRasterHeader) check(e error) {
	if e != nil {
		panic(e)
	}
}

func (r *idrisiRaster) deleteFiles() (err error) {
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
