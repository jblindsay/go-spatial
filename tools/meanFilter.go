// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Aug. 2015.

package tools

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

type MeanFilter struct {
	inputFile   string
	outputFile  string
	filterSizeX int
	filterSizeY int
	toolManager *PluginToolManager
}

func (this *MeanFilter) GetName() string {
	s := "MeanFilter"
	return getFormattedToolName(s)
}

// Returns a short description of the tool.
func (this *MeanFilter) GetDescription() string {
	s := "Performs a mean filtering operation on a raster"
	return getFormattedToolDescription(s)
}

func (this *MeanFilter) GetHelpDocumentation() string {
	ret := ""
	return ret
}

func (this *MeanFilter) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *MeanFilter) GetArgDescriptions() [][]string {
	numArgs := 4

	ret := make([][]string, numArgs)
	for i := range ret {
		ret[i] = make([]string, 3)
	}
	ret[0][0] = "InputFile"
	ret[0][1] = "string"
	ret[0][2] = "The input DEM File name, with directory and file extension"

	ret[1][0] = "OutputFile"
	ret[1][1] = "string"
	ret[1][2] = "The output filename, with directory and file extension"

	ret[2][0] = "FilterSizeX"
	ret[2][1] = "integer"
	ret[2][2] = "Filter size in the X direction"

	ret[3][0] = "FilterSizeY"
	ret[3][1] = "integer"
	ret[3][2] = "Filter size in the Y direction"

	return ret
}

func (this *MeanFilter) ParseArguments(args []string) {
	if len(args) != 4 {
		panic("The wrong number of arguments have been provided.")
	}
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

	this.filterSizeX = 3
	if len(strings.TrimSpace(args[2])) > 0 && args[2] != "not specified" {
		var err error
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(args[2]), 0, 0); err != nil {
			println(err)
		} else {
			this.filterSizeX = int(val)
		}
	}

	this.filterSizeY = this.filterSizeX
	if len(strings.TrimSpace(args[3])) > 0 && args[3] != "not specified" {
		var err error
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(args[3]), 0, 0); err != nil {
			println(err)
		} else {
			this.filterSizeY = int(val)
		}
	}

	this.Run()
}

func (this *MeanFilter) CollectArguments() {
	consolereader := bufio.NewReader(os.Stdin)

	// get the input file name
	fmt.Printf("\nEnter the raster file name (incl. file extension): ")
	inputFile, err := consolereader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	inputFile = strings.TrimSpace(inputFile)
	if !strings.Contains(inputFile, pathSep) {
		inputFile = this.toolManager.workingDirectory + inputFile
	}
	this.inputFile = inputFile
	// see if the file exists
	if _, err := os.Stat(this.inputFile); os.IsNotExist(err) {
		fmt.Printf("no such file or directory: %s\n", this.inputFile)
		return
	}

	// get the output file name
	fmt.Printf("\nEnter the output file name (incl. file extension): ")
	outputFile, err := consolereader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
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

	fmt.Printf("\nFilter size in X direction (grid cells): ")
	this.filterSizeX = 3
	filterSizeXStr, err := consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}
	if len(strings.TrimSpace(filterSizeXStr)) > 0 {
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(filterSizeXStr), 0, 0); err != nil {
			fmt.Println(err)
		} else {
			this.filterSizeX = int(val)
		}
	}

	fmt.Printf("\nFilter size in X direction (grid cells): ")
	this.filterSizeY = this.filterSizeX
	filterSizeYStr, err := consolereader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	if len(strings.TrimSpace(filterSizeYStr)) > 0 {
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(filterSizeYStr), 0, 0); err != nil {
			fmt.Println(err)
		} else {
			this.filterSizeY = int(val)
		}
	}

	this.Run()
}

func (this *MeanFilter) Run() {
	start1 := time.Now()

	var progress, oldProgress int

	fmt.Println("Reading raster data...")
	rin, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		fmt.Println(err.Error())
	}

	start2 := time.Now()

	rows := rin.Rows
	columns := rin.Columns
	rowsLessOne := rows - 1
	nodata := rin.NoDataValue
	inConfig := rin.GetRasterConfig()
	// rin.GetRasterConfig()

	// create the output raster
	config := raster.NewDefaultRasterConfig()
	config.PreferredPalette = inConfig.PreferredPalette
	config.DataType = raster.DT_FLOAT32
	config.NoDataValue = nodata
	config.InitialValue = nodata
	config.CoordinateRefSystemWKT = inConfig.CoordinateRefSystemWKT
	config.EPSGCode = inConfig.EPSGCode
	rout, err := raster.CreateNewRaster(this.outputFile, rows, columns,
		rin.North, rin.South, rin.East, rin.West, config)
	if err != nil {
		fmt.Println("Failed to write raster")
		return
	}

	numCPUs := runtime.NumCPU()
	c1 := make(chan int)
	runtime.GOMAXPROCS(numCPUs)
	var wg sync.WaitGroup

	// calculate hillshade
	// fmt.Printf("\r                                                    ")
	fmt.Printf("Progress: %v%%\n", 0)
	startingRow := 0
	var rowBlockSize int = rows / numCPUs

	numCells := 0

	k := 0
	for startingRow < rows {
		endingRow := startingRow + rowBlockSize
		if endingRow >= rows {
			endingRow = rows - 1
		}
		wg.Add(1)
		go func(rowSt, rowEnd, k int) {
			defer wg.Done()
			var z, zN float64
			numCellsInFilter := this.filterSizeX * this.filterSizeY
			halfFilterX := int(math.Floor(float64(this.filterSizeX) / 2.0))
			halfFilterY := int(math.Floor(float64(this.filterSizeY) / 2.0))
			dX := make([]int, numCellsInFilter, numCellsInFilter)
			dY := make([]int, numCellsInFilter, numCellsInFilter)
			i := 0
			for row := -halfFilterY; row <= halfFilterY; row++ {
				for col := -halfFilterX; col <= halfFilterX; col++ {
					dX[i] = col
					dY[i] = row
					i++
				}
			}
			for row := rowSt; row <= rowEnd; row++ {
				rowNumCells := 0
				floatData := make([]float64, columns)
				for col := 0; col < columns; col++ {
					z = rin.Value(row, col)
					if z != nodata {
						total := 0.0
						numNeighbours := 0.0
						for n := 0; n < numCellsInFilter; n++ {
							zN = rin.Value(row+dY[n], col+dX[n])
							if zN != nodata {
								total += zN
								numNeighbours += 1.0
							}
						}

						if numNeighbours > 0 {
							floatData[col] = total / numNeighbours
							rowNumCells++
						}
					} else {
						floatData[col] = nodata
					}
				}
				rout.SetRowValues(row, floatData)
				c1 <- rowNumCells
			}
		}(startingRow, endingRow, k)
		startingRow = endingRow + 1
		k++
	}

	oldProgress = 0
	for rowsCompleted := 0; rowsCompleted < rows; rowsCompleted++ {
		// rowNumCells := <-c1
		numCells += <-c1
		progress = int(100.0 * float64(rowsCompleted) / float64(rowsLessOne))
		if progress != oldProgress {
			fmt.Printf("Progress: %v%%\n", progress)
			oldProgress = progress
		}
	}

	wg.Wait()

	// fmt.Printf("                                                           ")
	fmt.Println("Saving data...")

	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)

	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by Slope"))
	rout.Save()

	fmt.Println("Operation complete!")

	value := fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	fmt.Println(value)

	overallTime := time.Since(start1)
	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	fmt.Println(value)
}
