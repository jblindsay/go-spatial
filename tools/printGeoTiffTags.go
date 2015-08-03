// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Feb. 2015.

package tools

import (
	"bufio"
	"gospatial/geospatialfiles/raster"
	"os"
	"strings"
)

type PrintGeoTiffTags struct {
	inputFile   string
	toolManager *PluginToolManager
}

func (this *PrintGeoTiffTags) GetName() string {
	s := "PrintGeoTiffTags"
	return getFormattedToolName(s)
}

func (this *PrintGeoTiffTags) GetDescription() string {
	s := "Prints a GeoTiff's tags"
	return getFormattedToolDescription(s)
}

func (this *PrintGeoTiffTags) GetHelpDocumentation() string {
	ret := "This tool prints the tags contained within a GeoTIFF file."
	return ret
}

func (this *PrintGeoTiffTags) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *PrintGeoTiffTags) GetArgDescriptions() [][]string {
	numArgs := 1

	ret := make([][]string, numArgs)
	for i := range ret {
		ret[i] = make([]string, 3)
	}
	ret[0][0] = "InputFile"
	ret[0][1] = "string"
	ret[0][2] = "The input GeoTiff file name"

	return ret
}

func (this *PrintGeoTiffTags) ParseArguments(args []string) {
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

	this.Run()
}

func (this *PrintGeoTiffTags) CollectArguments() {
	consolereader := bufio.NewReader(os.Stdin)

	// get the input file name
	print("Enter the  file name (incl. file extension): ")
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

	this.Run()
}

func (this *PrintGeoTiffTags) Run() {

	rasterType, err := raster.DetermineRasterFormat(this.inputFile)
	if rasterType != raster.RT_GeoTiff || err != nil {
		println("The input file is not of a GeoTIFF format.")
		return
	}

	input, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		println(err.Error())
	}

	tagInfo := input.GetMetadataEntries()
	if len(tagInfo) > 0 {
		println(tagInfo[0])
	} else {
		println("Error reading metadata entries.")
	}
}
