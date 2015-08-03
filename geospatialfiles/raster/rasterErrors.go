// Copyright 2014 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// Originally created by John Lindsay, Nov. 2014.

package raster

import "errors"

var UnsupportedRasterFormatError = errors.New("Unsupported raster format.")
var MultipleRasterFormatError = errors.New("There are multiple possible raster formats for this file.")
var FileReadingError = errors.New("An error occurred while reading the data file.")
var FileWritingError = errors.New("An error occurred while writing the data file.")
var FileOpeningError = errors.New("An error occurred while opening the data file.")
var RasterInitializationError = errors.New("An error occurred while initializing the raster.")
var FileDeletingError = errors.New("There were problems deleting the file.")
var FileDoesNotExistError = errors.New("The file does not exist.")
var DataSetError = errors.New("An error occurred while setting the data.")
var FileIsNotProperlyFormated = errors.New("The file does not appear to be properly formated")
