// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Feb. 2015.

package tools

import (
	"bufio"
	"fmt"
	"gospatial/geospatialfiles/raster"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Quantiles struct {
	inputFile   string
	outputFile  string
	numBins     int
	toolManager *PluginToolManager
}

func (this *Quantiles) GetName() string {
	s := "Quantiles"
	return getFormattedToolName(s)
}

// Returns a short description of the tool.
func (this *Quantiles) GetDescription() string {
	s := "Tranforms raster values into quantiles"
	return getFormattedToolDescription(s)
}

func (this *Quantiles) GetHelpDocumentation() string {
	ret := ""
	return ret
}

func (this *Quantiles) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *Quantiles) GetArgDescriptions() [][]string {
	numArgs := 3

	ret := make([][]string, numArgs)
	for i := range ret {
		ret[i] = make([]string, 3)
	}
	ret[0][0] = "InputFile"
	ret[0][1] = "string"
	ret[0][2] = "The input File name, with directory and file extension"

	ret[1][0] = "OutputFile"
	ret[1][1] = "string"
	ret[1][2] = "The output filename, with directory and file extension"

	ret[2][0] = "NumBins"
	ret[2][1] = "int"
	ret[2][2] = "The number of bins used to calculate the histogram"

	return ret
}

func (this *Quantiles) ParseArguments(args []string) {
	inputFile := args[0]
	inputFile = strings.TrimSpace(inputFile)
	if !strings.Contains(inputFile, pathSep) {
		inputFile = this.toolManager.workingDirectory + inputFile
	}
	this.inputFile = inputFile
	// see if the file exists
	if _, err := os.Stat(this.inputFile); os.IsNotExist(err) {
		printf("no such file or directory: %s\n", this.inputFile)
		return
	}
	outputFile := args[1]
	outputFile = strings.TrimSpace(outputFile)
	if !strings.Contains(outputFile, pathSep) {
		outputFile = this.toolManager.workingDirectory + outputFile
	}
	rasterType, err := raster.DetermineRasterFormat(outputFile)
	if rasterType == raster.RT_UnknownRaster || err == raster.UnsupportedRasterFormatError {
		outputFile = outputFile + ".tif" // default to a geotiff
	}
	this.outputFile = outputFile

	this.numBins = 1
	if len(strings.TrimSpace(args[2])) > 0 && args[2] != "not specified" {
		var err error
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(args[2]), 0, 0); err != nil {
			println(err)
		} else {
			this.numBins = int(val)
		}
	}

	this.Run()
}

func (this *Quantiles) CollectArguments() {
	consolereader := bufio.NewReader(os.Stdin)

	// get the input file name
	print("Enter the raster file name (incl. file extension): ")
	inputFile, err := consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}
	inputFile = strings.TrimSpace(inputFile)
	if !strings.Contains(inputFile, pathSep) {
		inputFile = this.toolManager.workingDirectory + inputFile
	}
	this.inputFile = inputFile
	// see if the file exists
	if _, err := os.Stat(this.inputFile); os.IsNotExist(err) {
		printf("no such file or directory: %s\n", this.inputFile)
		return
	}

	// get the output file name
	print("Enter the output file name (incl. file extension): ")
	outputFile, err := consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}
	outputFile = strings.TrimSpace(outputFile)
	if !strings.Contains(outputFile, pathSep) {
		outputFile = this.toolManager.workingDirectory + outputFile
	}
	rasterType, err := raster.DetermineRasterFormat(outputFile)
	if rasterType == raster.RT_UnknownRaster || err == raster.UnsupportedRasterFormatError {
		outputFile = outputFile + ".tif" // default to a geotiff
	}
	this.outputFile = outputFile

	print("Number of histogram bins: ")
	this.numBins = 1
	str, err := consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}

	if len(strings.TrimSpace(str)) > 0 {
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(str), 0, 0); err != nil {
			println(err)
		} else {
			this.numBins = int(val)
		}
	}

	this.Run()
}

func (this *Quantiles) Run() {
	start1 := time.Now()

	var progress, oldProgress, col, row, i, bin int
	var z float64

	println("Reading raster data...")
	rin, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		println(err.Error())
	}

	start2 := time.Now()

	rows := rin.Rows
	columns := rin.Columns
	rowsLessOne := rows - 1
	nodata := rin.NoDataValue
	inConfig := rin.GetRasterConfig()
	minValue := rin.GetMinimumValue()
	maxValue := rin.GetMaximumValue()
	valueRange := math.Ceil(maxValue - minValue)

	println("Calculating quantiles...")

	highResNumBins := 10000
	highResBinSize := valueRange / float64(highResNumBins)

	primaryHisto := make([]int, highResNumBins)
	numValidCells := 0
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			z = rin.Value(row, col)
			if z != nodata {
				bin = int(math.Floor((z - minValue) / highResBinSize))
				if bin >= highResNumBins {
					bin = highResNumBins - 1
				}
				primaryHisto[bin]++
				numValidCells++
			}
		}
	}

	for i = 1; i < highResNumBins; i++ {
		primaryHisto[i] += primaryHisto[i-1]
	}

	cdf := make([]float64, highResNumBins)
	for i = 0; i < highResNumBins; i++ {
		cdf[i] = 100.0 * float64(primaryHisto[i]) / float64(numValidCells)
	}

	quantileProportion := 100.0 / float64(this.numBins)

	for i = 0; i < highResNumBins; i++ {
		primaryHisto[i] = int(math.Floor(cdf[i] / quantileProportion))
		if primaryHisto[i] == this.numBins {
			primaryHisto[i] = this.numBins - 1
		}
	}

	// create the output raster
	config := raster.NewDefaultRasterConfig()
	config.PreferredPalette = inConfig.PreferredPalette
	config.DataType = raster.DT_INT16
	config.NoDataValue = nodata
	config.InitialValue = nodata
	config.CoordinateRefSystemWKT = inConfig.CoordinateRefSystemWKT
	config.EPSGCode = inConfig.EPSGCode
	rout, err := raster.CreateNewRaster(this.outputFile, rows, columns,
		rin.North, rin.South, rin.East, rin.West, config)
	if err != nil {
		println("Failed to write raster")
		return
	}

	printf("\r                                                           ")

	oldProgress = -1
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			z = rin.Value(row, col)
			if z != nodata {
				i = int(math.Floor((z - minValue) / highResBinSize))
				if i >= highResNumBins {
					i = highResNumBins - 1
				}
				bin = primaryHisto[i]

				rout.SetValue(row, col, float64(bin+1))
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress != oldProgress {
			printf("\rProgress: %v%%", progress)
			oldProgress = progress
		}
	}

	printf("\r                                                           ")
	printf("\rSaving data...\n")

	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by Quantiles with %v bins", this.numBins))
	rout.Save()

	println("Operation complete!")

	value := fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	println(value)

	overallTime := time.Since(start1)
	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}
