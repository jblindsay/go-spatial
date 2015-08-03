// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Feb. 2015.

package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

type FillSmallNodataHoles struct {
	inputFile   string
	outputFile  string
	toolManager *PluginToolManager
}

func (this *FillSmallNodataHoles) GetName() string {
	s := "FillSmallNodataHoles"
	return getFormattedToolName(s)
}

func (this *FillSmallNodataHoles) GetDescription() string {
	s := "Fills small nodata holes in a raster"
	return getFormattedToolDescription(s)
}

func (this *FillSmallNodataHoles) GetHelpDocumentation() string {
	ret := ""
	return ret
}

func (this *FillSmallNodataHoles) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *FillSmallNodataHoles) GetArgDescriptions() [][]string {
	numArgs := 2

	ret := make([][]string, numArgs)
	for i := range ret {
		ret[i] = make([]string, 3)
	}
	ret[0][0] = "InputDEM"
	ret[0][1] = "string"
	ret[0][2] = "The input DEM name, with directory and file extension"

	ret[1][0] = "OutputFile"
	ret[1][1] = "string"
	ret[1][2] = "The output filename, with directory and file extension"

	return ret
}

func (this *FillSmallNodataHoles) ParseArguments(args []string) {
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

	this.Run()
}

func (this *FillSmallNodataHoles) CollectArguments() {
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

	this.Run()
}

func (this *FillSmallNodataHoles) Run() {
	start1 := time.Now()

	var progress, oldProgress, col, row int
	var z, zN1, zN2 float64

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

	// create the output raster
	config := raster.NewDefaultRasterConfig()
	config.PreferredPalette = inConfig.PreferredPalette
	config.DataType = inConfig.DataType
	config.NoDataValue = nodata
	config.InitialValue = nodata
	config.CoordinateRefSystemWKT = inConfig.CoordinateRefSystemWKT
	config.EPSGCode = inConfig.EPSGCode
	config.DisplayMinimum = inConfig.DisplayMinimum
	config.DisplayMaximum = inConfig.DisplayMaximum
	rout, err := raster.CreateNewRaster(this.outputFile, rows, columns,
		rin.North, rin.South, rin.East, rin.West, config)
	if err != nil {
		println("Failed to write raster")
		return
	}

	printf("\r                                                           ")

	oldProgress = -1
	for row = 1; row < rows-1; row++ {
		for col = 0; col < columns; col++ {
			z = rin.Value(row, col)
			if z == nodata {
				zN1 = rin.Value(row-1, col)
				zN2 = rin.Value(row+1, col)
				if zN1 != nodata && zN2 != nodata {
					rout.SetValue(row, col, (zN1+zN2)/2.0)
				}
			} else {
				rout.SetValue(row, col, z)
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress != oldProgress {
			printf("\rProgress (Loop 1 of 2): %v%%", progress)
			oldProgress = progress
		}
	}

	oldProgress = -1
	for row = 0; row < rows; row++ {
		for col = 1; col < columns-1; col++ {
			z = rout.Value(row, col)
			if z == nodata {
				zN1 = rout.Value(row, col-1)
				zN2 = rout.Value(row, col+1)
				if zN1 != nodata && zN2 != nodata {
					rout.SetValue(row, col, (zN1+zN2)/2.0)
				}
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress != oldProgress {
			printf("\rProgress (Loop 2 of 2): %v%%", progress)
			oldProgress = progress
		}
	}

	printf("\r                                                           ")
	printf("\rSaving data...\n")

	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by FillSmallNodataHoles"))
	rout.Save()

	println("Operation complete!")

	value := fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	println(value)

	overallTime := time.Since(start1)
	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}
