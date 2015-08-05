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

// Used to manipulate an Whitebox raster (.dep) file.
type whiteboxRaster struct {
	dataFile     string
	data         []float64
	header       whiteboxRasterHeader
	minimumValue float64
	maximumValue float64
	config       *RasterConfig
}

func (r *whiteboxRaster) InitializeRaster(fileName string,
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
	r.config.RasterFormat = RT_WhiteboxRaster

	// set the file names; if they exist, delete them
	// sort out the names of the header and data files
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == ".tas" {
		r.dataFile = fileName
		r.header.fileName = strings.Replace(fileName, ext, ".dep", -1)
	} else if ext == ".dep" {
		r.header.fileName = fileName
		r.dataFile = strings.Replace(fileName, ext, ".tas", -1)
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

// Retrieve the data file name (.tas) of this Whitebox raster file.
func (r *whiteboxRaster) FileName() string {
	return r.dataFile
}

// Set the data file name (.tas) of this Whitebox raster file.
func (r *whiteboxRaster) SetFileName(value string) (err error) {
	r.config = NewDefaultRasterConfig()

	// sort out the names of the header and data files
	ext := strings.ToLower(filepath.Ext(value))
	if ext == ".tas" {
		r.dataFile = value
		r.header.fileName = strings.Replace(value, ext, ".dep", -1)
	} else if ext == ".dep" {
		r.header.fileName = value
		r.dataFile = strings.Replace(value, ext, ".tas", -1)
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
	r.config.RasterFormat = RT_WhiteboxRaster

	return nil
}

// Retrieve the RasterType of this Raster.
func (r *whiteboxRaster) RasterType() RasterType {
	return RT_WhiteboxRaster
}

// Retrieve the number of rows this binary raster file.
func (r *whiteboxRaster) Rows() int {
	return r.header.rows
}

// Sets the number of rows of this binary raster file.
func (r *whiteboxRaster) SetRows(value int) {
	r.header.rows = value
}

// Retrieve the number of columns of this binary raster file.
func (r *whiteboxRaster) Columns() int {
	return r.header.columns
}

// Sets the number of columns of this binary raster file.
func (r *whiteboxRaster) SetColumns(value int) {
	r.header.columns = value
}

// Retrieve the raster's northern edge's coordinate
func (r *whiteboxRaster) North() float64 {
	return r.header.north
}

// Retrieve the raster's southern edge's coordinate
func (r *whiteboxRaster) South() float64 {
	return r.header.south
}

// Retrieve the raster's eastern edge's coordinate
func (r *whiteboxRaster) East() float64 {
	return r.header.east
}

// Retrieve the raster's western edge's coordinate
func (r *whiteboxRaster) West() float64 {
	return r.header.west
}

// Retrieve the raster's minimum value
func (r *whiteboxRaster) MinimumValue() float64 {
	if r.minimumValue == math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.minimumValue
}

// Retrieve the raster's minimum value
func (r *whiteboxRaster) MaximumValue() float64 {
	if r.maximumValue == -math.MaxFloat64 {
		r.minimumValue, r.maximumValue = r.findMinAndMaxVals()
	}
	return r.maximumValue
}

func (r *whiteboxRaster) findMinAndMaxVals() (minVal float64, maxVal float64) {
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
func (r *whiteboxRaster) SetRasterConfig(value *RasterConfig) {
	r.config = value
}

// Retrieves the raster config
func (r *whiteboxRaster) GetRasterConfig() *RasterConfig {
	return r.config
}

// Retrieve the NoData value used by this binary raster file.
func (r *whiteboxRaster) NoData() float64 {
	return r.header.nodata
}

// Sets the NoData value used by this binary raster file.
func (r *whiteboxRaster) SetNoData(value float64) {
	r.header.nodata = value
	r.config.NoDataValue = value
}

// Retrieve the byte order used by this binary raster file.
func (r *whiteboxRaster) ByteOrder() binary.ByteOrder {
	return r.config.ByteOrder
}

// Sets the byte order used by this binary raster file.
func (r *whiteboxRaster) SetByteOrder(value binary.ByteOrder) {
	r.config.ByteOrder = value
}

// Retrieves the metadata for this raster
func (r *whiteboxRaster) MetadataEntries() []string {
	return r.config.MetadataEntries
}

// Adds a metadata entry to this raster
func (r *whiteboxRaster) AddMetadataEntry(value string) {
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
func (r *whiteboxRaster) Data() ([]float64, error) {
	if len(r.data) == 0 {
		r.ReadFile()
	}
	return r.data, nil
}

// Sets the data from a slice of float64 values
func (r *whiteboxRaster) SetData(values []float64) {
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
func (r *whiteboxRaster) Value(index int) float64 {
	return r.data[index]
}

// Sets the value of index within data
func (r *whiteboxRaster) SetValue(index int, value float64) {
	r.data[index] = value
}

//// Returns the value within ColorData
//func (r *whiteboxRaster) GetColor(index int) color.Color {
//	return r.colorData[index]
//}

//// Sets the value of index within ColorData
//func (r *whiteboxRaster) SetColor(index int, value color.Color) {
//	r.colorData[index] = value
//}

// Save the file
func (r *whiteboxRaster) Save() (err error) {
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
	case DT_FLOAT64:
		if err = binary.Write(buf, r.config.ByteOrder, r.data); err != nil {
			return FileWritingError
		}
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
	case DT_INT8:
		out := make([]int8, len(r.data))
		for i := 0; i < len(r.data); i++ {
			out[i] = int8(r.data[i])
		}
		if err = binary.Write(buf, r.config.ByteOrder, out); err != nil {
			return FileWritingError
		}
	default:
		return FileWritingError
	}
	w.Write(buf.Bytes())
	w.Flush()
	return nil
}

// Reads the file
func (r *whiteboxRaster) ReadFile() error {
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
	case DT_FLOAT64:
		err = binary.Read(buf, r.config.ByteOrder, &r.data)
		r.check(err)
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
	case DT_INT8:
		nativeData := make([]int8, r.header.numCells)
		err = binary.Read(buf, r.config.ByteOrder, &nativeData)
		r.check(err)
		for i, value := range nativeData {
			r.data[i] = float64(value)
		}
		nativeData = nil
	default:
		return FileReadingError
	}

	return nil
}

type whiteboxRasterHeader struct {
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

func (r *whiteboxRaster) readHeaderFile() error {
	// read the header file
	if r.header.fileName == "" {
		return errors.New("Whitebox GAT raster header file not set properly.")
	}
	content, err := ioutil.ReadFile(r.header.fileName)
	r.check(err)
	str := strings.Replace(string(content), "\r\n", "\n", -1)
	lines := strings.Split(str, "\n")
	for a := 0; a < len(lines); a++ {
		str = strings.ToLower(lines[a])
		s := strings.Split(lines[a], "\t")
		if strings.Contains(str, "min:") && !strings.Contains(str, "display") && !strings.Contains(str, "metadata entry") {
			r.minimumValue, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "max:") && !strings.Contains(str, "display") && !strings.Contains(str, "metadata entry") {
			r.maximumValue, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "display min") && !strings.Contains(str, "metadata entry") {
			r.config.DisplayMinimum, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "display max") && !strings.Contains(str, "metadata entry") {
			r.config.DisplayMaximum, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "north") && !strings.Contains(str, "metadata entry") {
			r.header.north, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "south") && !strings.Contains(str, "metadata entry") {
			r.header.south, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "east") && !strings.Contains(str, "metadata entry") {
			r.header.east, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "west") && !strings.Contains(str, "metadata entry") {
			r.header.west, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "cols") && !strings.Contains(str, "metadata entry") {
			r.header.columns, err = strconv.Atoi(s[len(s)-1])
			r.check(err)
		} else if strings.Contains(str, "rows") && !strings.Contains(str, "metadata entry") {
			r.header.rows, err = strconv.Atoi(s[len(s)-1])
			r.check(err)
		} else if strings.Contains(str, "stacks") && !strings.Contains(str, "metadata entry") {
			r.config.NumberOfBands, err = strconv.Atoi(s[len(s)-1])
			r.check(err)
		} else if strings.Contains(str, "data type") && !strings.Contains(str, "metadata entry") {
			dt := strings.ToLower(strings.TrimSpace(s[len(s)-1]))
			if strings.Contains(dt, "double") {
				r.config.DataType = DT_FLOAT64
			} else if strings.Contains(dt, "float") {
				r.config.DataType = DT_FLOAT32
			} else if strings.Contains(dt, "int") {
				r.config.DataType = DT_INT16
			} else { // byte
				r.config.DataType = DT_INT8
			}
		} else if strings.Contains(str, "data scale") && !strings.Contains(str, "metadata entry") {
			str2 := strings.ToLower(strings.TrimSpace(s[len(s)-1]))
			if str2 == "continuous" {
				r.config.PhotometricInterpretation = 0
			} else if str2 == "categorical" {
				r.config.PhotometricInterpretation = 1
			} else if str2 == "bool" {
				r.config.PhotometricInterpretation = 2
			} else if str2 == "rgb" {
				r.config.PhotometricInterpretation = 3
			} else { // continous is the default
				r.config.PhotometricInterpretation = 0
			}
		} else if strings.Contains(str, "z units") && !strings.Contains(str, "metadata entry") {
			r.config.ZUnits = strings.ToLower(strings.TrimSpace(s[len(s)-1]))
		} else if strings.Contains(str, "xy units") && !strings.Contains(str, "metadata entry") {
			r.config.XYUnits = strings.ToLower(strings.TrimSpace(s[len(s)-1]))
		} else if strings.Contains(str, "projection") && !strings.Contains(str, "metadata entry") {
			r.config.CoordinateRefSystemWKT = strings.TrimPrefix(lines[a], "Projection:\t")
		} else if strings.Contains(str, "preferred palette") && !strings.Contains(str, "metadata entry") {
			r.config.PreferredPalette = strings.ToLower(strings.TrimSpace(s[len(s)-1]))
		} else if strings.Contains(str, "byteorder") && !strings.Contains(str, "metadata entry") {
			if strings.Contains(strings.ToLower(s[len(s)-1]), "LITTLE_ENDIAN") {
				r.config.ByteOrder = binary.LittleEndian
			} else {
				r.config.ByteOrder = binary.BigEndian
			}
		} else if strings.Contains(str, "nodata") && !strings.Contains(str, "metadata entry") {
			r.header.nodata, err = strconv.ParseFloat(s[len(s)-1], 64)
			r.check(err)
		} else if strings.Contains(str, "metadata entry") {
			value := strings.TrimSpace(s[len(s)-1])
			value = strings.Replace(value, ";", ":", -1)
			r.AddMetadataEntry(value)
			//r.config.MetadataEntries = append(r.config.MetadataEntries, value)
		}
	}

	r.header.numCells = r.header.rows * r.header.columns

	return nil
}

func (r *whiteboxRaster) writeHeaderFile() (err error) {
	f, err := os.Create(r.header.fileName)
	r.check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	var str string

	r.minimumValue, r.maximumValue = r.findMinAndMaxVals()

	str = "Min:\t" + strconv.FormatFloat(r.minimumValue, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "Max:\t" + strconv.FormatFloat(r.maximumValue, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "North:\t" + strconv.FormatFloat(r.header.north, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "South:\t" + strconv.FormatFloat(r.header.south, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "East:\t" + strconv.FormatFloat(r.header.east, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "West:\t" + strconv.FormatFloat(r.header.west, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "Cols:\t" + strconv.Itoa(r.header.columns)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "Rows:\t" + strconv.Itoa(r.header.rows)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "Stacks:\t" + strconv.Itoa(r.config.NumberOfBands)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	switch r.config.DataType {
	case DT_FLOAT64:
		str = "Data Type:\tDOUBLE"
	case DT_INT16:
		str = "Data Type:\tINTEGER"
	case DT_INT8:
		str = "Data Type:\tBYTE"
	default:
		str = "Data Type:\tFLOAT"
	}

	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "Z Units:\t" + r.config.ZUnits
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "XY Units:\t" + r.config.XYUnits
	_, err = w.WriteString(str + "\n")
	r.check(err)

	if r.config.CoordinateRefSystemWKT == "" {
		r.config.CoordinateRefSystemWKT = "not specified"
	}
	str = "Projection:\t" + r.config.CoordinateRefSystemWKT
	_, err = w.WriteString(str + "\n")
	r.check(err)

	switch r.config.PhotometricInterpretation {
	case 0:
		str = "Data Scale:\tcontinuous"
		_, err = w.WriteString(str + "\n")
		r.check(err)
	case 1:
		str = "Data Scale:\tcategorical"
		_, err = w.WriteString(str + "\n")
		r.check(err)
	case 2:
		str = "Data Scale:\tboolean"
		_, err = w.WriteString(str + "\n")
		r.check(err)
	case 3:
		str = "Data Scale:\trgb"
		_, err = w.WriteString(str + "\n")
		r.check(err)
	}

	if r.config.DisplayMinimum == math.MaxFloat64 {
		r.config.DisplayMinimum = r.minimumValue
	}
	str = "Display Min:\t" + strconv.FormatFloat(r.config.DisplayMinimum, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	if r.config.DisplayMaximum == -math.MaxFloat64 {
		r.config.DisplayMaximum = r.maximumValue
	}
	str = "Display Max:\t" + strconv.FormatFloat(r.config.DisplayMaximum, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	if r.config.PreferredPalette == "not specified" {
		r.config.PreferredPalette = "grey.pal"
	}
	str = "Preferred Palette:\t" + r.config.PreferredPalette
	_, err = w.WriteString(str + "\n")
	r.check(err)

	str = "NoData:\t" + strconv.FormatFloat(r.header.nodata, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)
	if r.config.ByteOrder == binary.LittleEndian {
		str = "Byte Order:\tLITTLE_ENDIAN"
		_, err = w.WriteString(str + "\n")
		r.check(err)
	} else {
		str = "Byte Order:\tBIG_ENDIAN"
		_, err = w.WriteString(str + "\n")
		r.check(err)
	}
	str = "Palette Nonlinearity:\t" + strconv.FormatFloat(r.config.PaletteNonlinearity, 'f', -1, 64)
	_, err = w.WriteString(str + "\n")
	r.check(err)

	// write the metadata entries
	for _, value := range r.config.MetadataEntries {
		if len(strings.TrimSpace(value)) > 0 {
			str = "Metadata Entry:\t" + strings.Replace(value, ":", ";", -1)
			_, err = w.WriteString(str + "\n")
			r.check(err)
		}
	}

	w.Flush()
	return nil
}

func (r *whiteboxRaster) check(e error) {
	if e != nil {
		panic(e)
	}
}

func (h *whiteboxRasterHeader) check(e error) {
	if e != nil {
		panic(e)
	}
}

func (r *whiteboxRaster) deleteFiles() (err error) {
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
