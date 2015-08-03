// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Feb. 2015.

package tools

import (
	"bufio"
	"os"
	"strings"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

type Whitebox2GeoTiff struct {
	inputFile   string
	outputFile  string
	toolManager *PluginToolManager
}

func (this *Whitebox2GeoTiff) GetName() string {
	s := "Whitebox2GeoTiff"
	return getFormattedToolName(s)
}

func (this *Whitebox2GeoTiff) GetDescription() string {
	s := "Converts Whitebox GAT raster to GeoTiff"
	return getFormattedToolDescription(s)
}

func (this *Whitebox2GeoTiff) GetHelpDocumentation() string {
	ret := "This tool converts a Whitebox GAT raster to a GeoTiff format."
	return ret
}

func (this *Whitebox2GeoTiff) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *Whitebox2GeoTiff) GetArgDescriptions() [][]string {
	numArgs := 2

	ret := make([][]string, numArgs)
	for i := range ret {
		ret[i] = make([]string, 3)
	}
	ret[0][0] = "InputFile"
	ret[0][1] = "string"
	ret[0][2] = "The input Whitebox GAT file name"

	ret[1][0] = "OutputFile"
	ret[1][1] = "string"
	ret[1][2] = "The output GeoTiff file name"

	return ret
}

func (this *Whitebox2GeoTiff) ParseArguments(args []string) {
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
	this.outputFile = outputFile

	this.Run()
}

func (this *Whitebox2GeoTiff) CollectArguments() {
	consolereader := bufio.NewReader(os.Stdin)

	// get the input file name
	print("Enter the input Whitebox raster file name (incl. file extension): ")
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
	print("Enter the output GeoTiff raster file name (incl. file extension): ")
	outputFile, err := consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}
	outputFile = strings.TrimSpace(outputFile)
	if !strings.Contains(outputFile, pathSep) {
		outputFile = this.toolManager.workingDirectory + outputFile
	}
	this.outputFile = outputFile

	this.Run()
}

func (this *Whitebox2GeoTiff) Run() {

	// check that the input raster is in Whitebox GAT format
	rasterType, err := raster.DetermineRasterFormat(this.inputFile)
	if rasterType != raster.RT_WhiteboxRaster || err != nil {
		println("The input file is not of a Whitebox GAT format.")
		return
	}

	input, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		println(err.Error())
	}

	// get the input config
	inConfig := input.GetRasterConfig()

	// get the number of rows and columns
	rows := input.Rows
	columns := input.Columns
	rowsLessOne := rows - 1
	inNodata := input.NoDataValue

	// check that the specified output file is in GeoTiff format
	rasterType, err = raster.DetermineRasterFormat(this.outputFile)
	if rasterType != raster.RT_GeoTiff || err != nil {
		println("Warning: The specified output file name is not of a GeoTIFF format.\nThe file name has been modified")
		index := strings.LastIndex(this.outputFile, ".")
		extension := this.outputFile[index:len(this.outputFile)]
		newFileName := strings.Replace(this.outputFile, extension, ".tif", -1)
		this.outputFile = newFileName
	}

	// output the data
	outConfig := raster.NewDefaultRasterConfig()
	outConfig.DataType = inConfig.DataType
	outConfig.EPSGCode = inConfig.EPSGCode
	//outConfig.NoDataValue = inConfig.NoDataValue
	outConfig.CoordinateRefSystemWKT = inConfig.CoordinateRefSystemWKT
	output, err := raster.CreateNewRaster(this.outputFile, input.Rows, input.Columns,
		input.North, input.South, input.East, input.West, outConfig)
	outNodata := output.NoDataValue
	if err != nil {
		println(err.Error())
	}

	var progress, oldProgress int
	var z float64
	oldProgress = -1
	for row := 0; row < rows; row++ {
		for col := 0; col < columns; col++ {
			z = input.Value(row, col)
			if z != inNodata {
				output.SetValue(row, col, z)
			} else {
				output.SetValue(row, col, outNodata)
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress != oldProgress {
			printf("\rProgress: %v%%", progress)
			oldProgress = progress
		}
	}
	output.Save()
	println("\nOperation complete!")
}
