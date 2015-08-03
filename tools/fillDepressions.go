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

type FillDepressions struct {
	inputFile   string
	outputFile  string
	fixFlats    bool
	toolManager *PluginToolManager
}

func (this *FillDepressions) GetName() string {
	s := "FillDepressions"
	return getFormattedToolName(s)
}

func (this *FillDepressions) GetDescription() string {
	s := "Removes depressions in DEMs using filling"
	return getFormattedToolDescription(s)
}

func (this *FillDepressions) GetHelpDocumentation() string {
	ret := "This tool is used to remove the sinks (i.e. topographic depressions and flat areas) from digital elevation models (DEMs) using an efficient depression filling method. Note that the BreachDepressions tool is the preferred method of creating a depressionless DEM."
	return ret
}

func (this *FillDepressions) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *FillDepressions) GetArgDescriptions() [][]string {
	numArgs := 3

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

	ret[2][0] = "FixFlats"
	ret[2][1] = "bool"
	ret[2][2] = "Should the resulting flat areas be fixed?"

	return ret
}

func (this *FillDepressions) ParseArguments(args []string) {
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

	this.fixFlats = false
	if len(strings.TrimSpace(args[2])) > 0 && args[2] != "not specified" {
		var err error
		if this.fixFlats, err = strconv.ParseBool(strings.TrimSpace(args[2])); err != nil {
			this.fixFlats = false
			println(err)
		}
	} else {
		this.fixFlats = false
	}
	this.Run()
}

func (this *FillDepressions) CollectArguments() {
	consolereader := bufio.NewReader(os.Stdin)

	// get the input file name
	print("Enter the DEM file name (incl. file extension): ")
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

	// get the fixflats argument
	print("Fix the resulting flat areas (T or F)? ")
	fixFlatsStr, err := consolereader.ReadString('\n')
	if err != nil {
		this.fixFlats = false
		println(err)
	}

	if len(strings.TrimSpace(fixFlatsStr)) > 0 {
		if this.fixFlats, err = strconv.ParseBool(strings.TrimSpace(fixFlatsStr)); err != nil {
			this.fixFlats = false
			println(err)
		}
	} else {
		this.fixFlats = false
	}

	this.Run()
}

func (this *FillDepressions) Run() {

	if this.toolManager.BenchMode {
		benchmarkFillDepressions(this)
		return
	}

	start1 := time.Now()

	var progress, oldProgress, col, row, i, n int
	var colN, rowN, flatindex int
	numSolvedCells := 0
	var z, zN float64
	var gc gridCell
	var p int64
	var isEdgeCell bool
	dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
	dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}

	println("Reading DEM data...")
	dem, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		println(err.Error())
	}
	rows := dem.Rows
	columns := dem.Columns
	rowsLessOne := rows - 1
	numCellsTotal := rows * columns
	nodata := dem.NoDataValue
	demConfig := dem.GetRasterConfig()
	paletteName := demConfig.PreferredPalette

	// output the data
	// make a copy of the dem's raster configuration
	//config := dem.GetRasterConfig()
	config := raster.NewDefaultRasterConfig()
	config.PreferredPalette = paletteName
	config.DataType = raster.DT_FLOAT32
	config.NoDataValue = nodata
	config.InitialValue = nodata
	displayMin := demConfig.DisplayMinimum
	displayMax := demConfig.DisplayMaximum
	config.DisplayMinimum = displayMin
	config.DisplayMaximum = displayMax
	config.CoordinateRefSystemWKT = demConfig.CoordinateRefSystemWKT
	config.EPSGCode = demConfig.EPSGCode
	value := fmt.Sprintf("Created on %s\n", time.Now().Local())
	config.MetadataEntries = append(config.MetadataEntries, value)
	rout, err := raster.CreateNewRaster(this.outputFile, rows, columns,
		dem.North, dem.South, dem.East, dem.West, config)
	if err != nil {
		panic("Failed to write raster")
	}

	minVal := dem.GetMinimumValue()
	elevDigits := len(strconv.Itoa(int(dem.GetMaximumValue() - minVal)))
	elevMultiplier := math.Pow(10, float64(8-elevDigits))
	SMALL_NUM := 1 / elevMultiplier
	if !this.fixFlats {
		SMALL_NUM = 0
	}

	start2 := time.Now()

	// Fill the DEM.
	inQueue := make([][]bool, rows+2)

	for i = 0; i < rows+2; i++ {
		inQueue[i] = make([]bool, columns+2)
	}
	//inQueue := structures.Create2dBoolArray(rows+2, columns+2)

	// Reinitialize the priority queue and flow direction grid.
	numSolvedCells = 0

	//pq := make(PriorityQueue, 0)
	pq := NewPQueue()

	// find the pit cells and initialize the grids
	printf("\r                                                      ")
	printf("\rFilling DEM (1 of 2): %v%%", 0)
	oldProgress = 0
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			z = dem.Value(row, col)
			if z != nodata {
				//isPit = true
				isEdgeCell = false
				for n = 0; n < 8; n++ {
					zN = dem.Value(row+dY[n], col+dX[n])
					if zN == nodata {
						isEdgeCell = true
					} // else if zN < z {
					//	isPit = false
					//}
				}

				if isEdgeCell { //}&& isPit {
					gc = newGridCell(row, col, 0)
					p = int64(int64(zN*elevMultiplier) * 100000)
					//					item := &Item{
					//						value:    gc,
					//						priority: p,
					//					}
					//					heap.Push(&pq, item)
					pq.Push(gc, p)
					inQueue[row+1][col+1] = true
					rout.SetValue(row, col, z)
					numSolvedCells++
				}
			} else {
				numSolvedCells++
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress != oldProgress {
			printf("\rFilling DEM (1 of 2): %v%%", progress)
			oldProgress = progress
		}
	}

	//heap.Init(&pq)
	printf("\r                                                      ")
	oldProgress = -1
	for numSolvedCells < numCellsTotal { //pq.Len() > 0 {
		//item := heap.Pop(&pq).(*Item)
		//gc = item.value
		gc = pq.Pop()
		row = gc.row
		col = gc.column
		flatindex = gc.flatIndex
		z = rout.Value(row, col)
		for i = 0; i < 8; i++ {
			rowN = row + dY[i]
			colN = col + dX[i]
			zN = dem.Value(rowN, colN)
			if zN != nodata && !inQueue[rowN+1][colN+1] {
				n = 0
				if zN <= z {
					zN = z + SMALL_NUM
					n = flatindex + 1
				}
				numSolvedCells++
				rout.SetValue(rowN, colN, zN)
				gc = newGridCell(rowN, colN, n)
				p = int64(int64(zN*elevMultiplier)*100000 + (int64(n) % 100000))
				//				item = &Item{
				//					value:    gc,
				//					priority: p,
				//				}
				//				heap.Push(&pq, item)
				pq.Push(gc, p)
				inQueue[rowN+1][colN+1] = true
			}
		}
		progress = int(100.0 * numSolvedCells / numCellsTotal)
		if progress != oldProgress {
			printf("\rFilling DEM (2 of 2): %v%%", progress)
			oldProgress = progress
		}
	}

	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by FillDepressions tool"))
	config.DisplayMinimum = displayMin
	config.DisplayMaximum = displayMax
	rout.SetRasterConfig(config)
	rout.Save()

	println("\nOperation complete!")

	value = fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	println(value)

	overallTime := time.Since(start1)
	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}
