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
	"strings"
	"sync"
	"time"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

type Aspect struct {
	inputFile   string
	outputFile  string
	toolManager *PluginToolManager
}

func (this *Aspect) GetName() string {
	s := "Aspect"
	return getFormattedToolName(s)
}

// Returns a short description of the tool.
func (this *Aspect) GetDescription() string {
	s := "Calculates aspect from a DEM"
	return getFormattedToolDescription(s)
}

func (this *Aspect) GetHelpDocumentation() string {
	ret := ""
	return ret
}

func (this *Aspect) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *Aspect) GetArgDescriptions() [][]string {
	numArgs := 2

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

	return ret
}

func (this *Aspect) ParseArguments(args []string) {
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

func (this *Aspect) CollectArguments() {
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

func (this *Aspect) Run() {
	start1 := time.Now()

	var progress, oldProgress int

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
	gridRes := (rin.GetCellSizeX() + rin.GetCellSizeY()) / 2.0
	eightGridRes := 8 * gridRes
	const radToDeg float64 = 180.0 / math.Pi
	rin.GetRasterConfig()

	// create the output raster
	config := raster.NewDefaultRasterConfig()
	config.PreferredPalette = "circular_bw.pal"
	config.DataType = raster.DT_FLOAT32
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

	zConvFactor := 1.0
	if rin.IsInGeographicCoordinates() {
		// calculate a new z-conversion factor
		midLat := (rin.North - rin.South) / 2.0
		if midLat <= 90 && midLat >= -90 {
			zConvFactor = 1.0 / (113200 * math.Cos(math.Pi/180.0*midLat))
		}
	}

	numCPUs := runtime.NumCPU()
	c1 := make(chan bool)
	runtime.GOMAXPROCS(numCPUs)
	var wg sync.WaitGroup

	// calculate aspect
	printf("\r                                                    ")
	printf("\rProgress: %v%%", 0)
	//var numSolvedCells int = 0
	startingRow := 0
	var rowBlockSize int = rows / numCPUs

	k := 0
	for startingRow < rows {
		endingRow := startingRow + rowBlockSize
		if endingRow >= rows {
			endingRow = rows - 1
		}
		wg.Add(1)
		go func(rowSt, rowEnd, k int) {
			defer wg.Done()
			var z, zN, fy, fx, value float64
			dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
			dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
			N := [8]float64{}
			for row := rowSt; row <= rowEnd; row++ {
				floatData := make([]float64, columns)
				for col := 0; col < columns; col++ {
					z = rin.Value(row, col)
					if z != nodata {
						z = z * zConvFactor
						for n := 0; n < 8; n++ {
							zN = rin.Value(row+dY[n], col+dX[n])
							if zN != nodata {
								N[n] = zN * zConvFactor
							} else {
								N[n] = z
							}
						}

						fy = (N[6] - N[4] + 2*(N[7]-N[3]) + N[0] - N[2]) / eightGridRes
						fx = (N[2] - N[4] + 2*(N[1]-N[5]) + N[0] - N[6]) / eightGridRes

						if fx != 0 {
							value = 180 - math.Atan(fy/fx)*radToDeg + 90*(fx/math.Abs(fx))
							floatData[col] = value
						} else {
							floatData[col] = -1.0
						}
					} else {
						floatData[col] = nodata
					}
				}
				rout.SetRowValues(row, floatData)
				c1 <- true // row completed
			}
		}(startingRow, endingRow, k)
		startingRow = endingRow + 1
		k++
	}

	oldProgress = 0
	for rowsCompleted := 0; rowsCompleted < rows; rowsCompleted++ {
		<-c1 // a row has successfully completed
		progress = int(100.0 * float64(rowsCompleted) / float64(rowsLessOne))
		if progress != oldProgress {
			printf("\rProgress: %v%%", progress)
			oldProgress = progress
		}
	}

	wg.Wait()

	printf("\r                                                           ")
	printf("\rSaving data...\n")

	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by Slope"))
	rout.Save()

	println("Operation complete!")

	value := fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	println(value)

	overallTime := time.Since(start1)
	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}
