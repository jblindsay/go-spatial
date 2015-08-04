// Copyright 2014 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// Originally created by John Lindsay<jlindsay@uoguelph.ca>, Nov. 2014.

// Package raster provides support for reading and creating various common
// geospatial raster data formats.
package raster

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// RasterType is used to specify a data format of a raster file
type RasterType int

// Integer constants used to specify each of the supported raster formats
const (
	RT_UnknownRaster      RasterType = 0
	RT_ArcGisBinaryRaster RasterType = iota
	RT_ArcGisAsciiRaster
	RT_GeoTiff
	RT_WhiteboxRaster
	RT_GrassAsciiRaster
	RT_SurferAsciiRaster
	RT_SagaRaster
	RT_IdrisiRaster
)

var rasterTypeList = []string{
	"UnknownRaster",
	"ArcGisBinaryRaster",
	"ArcGisAsciiRaster",
	"GeoTiff",
	"WhiteboxRaster",
	"GrassAsciiRaster",
	"SurferAsciiRaster",
	"SagaRaster",
	"IdrisiRaster",
}

// String returns the English name of the RasterType ("ArcGisBinaryRaster", "ArcGisAsciiRaster", ...).
func (rt RasterType) String() string { return rasterTypeList[rt] }

var rasterExtensionList [][]string

func init() {
	rasterExtensionList = make([][]string, 0)
	rasterExtensionList = append(rasterExtensionList, []string{".*"})
	rasterExtensionList = append(rasterExtensionList, []string{".flt", ".hdr"})
	rasterExtensionList = append(rasterExtensionList, []string{".txt", ".asc"})
	rasterExtensionList = append(rasterExtensionList, []string{".tif", ".tiff"})
	rasterExtensionList = append(rasterExtensionList, []string{".tas", ".dep"})
	rasterExtensionList = append(rasterExtensionList, []string{".txt"})
	rasterExtensionList = append(rasterExtensionList, []string{".grd"})
	rasterExtensionList = append(rasterExtensionList, []string{".sdat", ".sgrd"})
	rasterExtensionList = append(rasterExtensionList, []string{".rst", ".rdc"})
}

// Returns a list of the file extensions associated with a particular raster format.
func (rt RasterType) GetExtensions() []string {
	return rasterExtensionList[0]
}

func IsSupportedRasterFileExtension(fileName string) (ret bool) {
	// see if it is a supported raster format
	ret = false
	fileExtension := strings.ToLower(filepath.Ext(fileName))

	for _, extensions := range rasterExtensionList {
		for _, ext := range extensions {
			if fileExtension == ext {
				ret = true
				break
			}
		}
	}
	return
}

// Attempts to determine the raster format from the filename.
func DetermineRasterFormat(fileName string) (rt RasterType, err error) {
	rt = RT_UnknownRaster

	// get a list of each of the raster formats that have
	// the same file extension as the filename.
	fileExtension := strings.ToLower(filepath.Ext(fileName))
	list := make([]RasterType, 0)
	for i, extensions := range rasterExtensionList {
		for _, ext := range extensions {
			if fileExtension == ext {
				list = append(list, RasterType(i))
			}
		}
	}

	numPossibleFormats := len(list)
	if numPossibleFormats == 0 {
		// could not find a corresponding supported format
		return rt, UnsupportedRasterFormatError
	} else if numPossibleFormats == 1 {
		// there is only one unique format it could be
		return list[0], nil
	} else {
		// conflict resolution

		// first see if it's an existing file
		if _, err := os.Stat(fileName); err == nil {
			if fileExtension == ".txt" {
				// read in the first six lines of the file
				contents := ""
				f, err := os.Open(fileName)
				if err != nil {
					return rt, FileOpeningError
				}
				defer f.Close()

				scanner := bufio.NewScanner(f)
				j := 0
				for scanner.Scan() {
					contents += strings.ToLower(scanner.Text()) + "\n"
					j++
					if j == 6 {
						break
					}
				}

				if strings.Contains(contents, "ncols") &&
					strings.Contains(contents, "nrows") &&
					strings.Contains(contents, "xll") &&
					strings.Contains(contents, "yll") &&
					strings.Contains(contents, "yll") {
					// it's an ArcGIS ASCII raster
					rt = RasterType(2)
					return rt, nil
				} else if strings.Contains(contents, "north") &&
					strings.Contains(contents, "south") &&
					strings.Contains(contents, "east") &&
					strings.Contains(contents, "west") &&
					strings.Contains(contents, "rows") &&
					strings.Contains(contents, "cols") {
					// it's a GRASS ASCII raster
					rt = RasterType(5)
					return rt, nil
				}
			}
		} else {
			// The file does not already exist so there is no way to tell what the
			// format should be uniquely. Just return the first entry of list along
			// with a warning that there are multiple possible formats
			return list[0], MultipleRasterFormatError
		}
	}
	return
}

// String returns the English name of the RasterType ("ArcGisBinaryRaster", "ArcGisAsciiRaster", ...).
func ListAllSupportedRasterFormats() []string {
	return rasterTypeList
}

func GetMapOfFormatsAndExtensions() map[string][]string {
	m := make(map[string][]string)
	for i, val := range rasterTypeList {
		m[val] = rasterExtensionList[i]
	}
	return m
}
