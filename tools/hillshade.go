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

type Hillshade struct {
	inputFile   string
	outputFile  string
	toolManager *PluginToolManager
}

func (this *Hillshade) GetName() string {
	s := "Hillshade"
	return getFormattedToolName(s)
}

// Returns a short description of the tool.
func (this *Hillshade) GetDescription() string {
	s := "Calculates a hillshade raster from a DEM"
	return getFormattedToolDescription(s)
}

func (this *Hillshade) GetHelpDocumentation() string {
	ret := ""
	return ret
}

func (this *Hillshade) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *Hillshade) GetArgDescriptions() [][]string {
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

func (this *Hillshade) ParseArguments(args []string) {
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

func (this *Hillshade) CollectArguments() {
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

func (this *Hillshade) Run() {
	start1 := time.Now()

	var progress, oldProgress int

	azimuth := (315.0 - 90.0) * DegToRad
	altitude := 30.0 * DegToRad
	sinTheta := math.Sin(altitude)
	cosTheta := math.Cos(altitude)

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
	config.PreferredPalette = "grey.pal"
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
	c1 := make(chan [256]int)
	c2 := make(chan int)
	runtime.GOMAXPROCS(numCPUs)
	var wg sync.WaitGroup

	// calculate hillshade
	printf("\r                                                    ")
	printf("\rProgress: %v%%", 0)
	startingRow := 0
	var rowBlockSize int = rows / numCPUs

	histo := [256]int{}
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
			var z, zN, fy, fx, value, tanSlope, aspect, term1, term2, term3 float64
			dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
			dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
			N := [8]float64{}
			for row := rowSt; row <= rowEnd; row++ {
				rowHisto := [256]int{}
				rowNumCells := 0
				floatData := make([]float64, columns)
				for col := 0; col < columns; col++ {
					z = rin.Value(row, col)
					if z != nodata {
						z *= zConvFactor
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
							tanSlope = math.Sqrt(fx*fx + fy*fy)
							aspect = (180 - math.Atan(fy/fx)*RadToDeg + 90*(fx/math.Abs(fx))) * DegToRad
							term1 = tanSlope / math.Sqrt(1+tanSlope*tanSlope)
							term2 = sinTheta / tanSlope
							term3 = cosTheta * math.Sin(azimuth-aspect)
							z = term1 * (term2 - term3)
						} else {
							z = 0.5
						}

						value = math.Floor(z * 255)
						if value < 0 {
							value = 0
						}
						floatData[col] = value
						rowHisto[int(value)]++
						rowNumCells++
					} else {
						floatData[col] = nodata
					}
				}
				rout.SetRowValues(row, floatData)
				c1 <- rowHisto // row completed
				c2 <- rowNumCells
			}
		}(startingRow, endingRow, k)
		startingRow = endingRow + 1
		k++
	}

	//rowHisto := [256]int64{}
	oldProgress = 0
	for rowsCompleted := 0; rowsCompleted < rows; rowsCompleted++ {
		rowHisto := <-c1 // a row has successfully completed
		for i := 0; i < 256; i++ {
			histo[i] += rowHisto[i]
		}
		rowNumCells := <-c2
		numCells += rowNumCells
		progress = int(100.0 * float64(rowsCompleted) / float64(rowsLessOne))
		if progress != oldProgress {
			printf("\rProgress: %v%%", progress)
			oldProgress = progress
		}
	}

	wg.Wait()

	// trim the display min and max values by 1%
	newMin := 0.0
	newMax := 0.0
	targetCellNum := int(float64(numCells) * 0.01)
	sum := 0
	for i := 0; i < 256; i++ {
		sum += histo[i]
		if sum >= targetCellNum {
			newMin = float64(i)
			break
		}
	}

	sum = 0
	for i := 255; i >= 0; i-- {
		sum += histo[i]
		if sum >= targetCellNum {
			newMax = float64(i)
			break
		}
	}

	printf("\r                                                           ")
	printf("\rSaving data...\n")

	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)

	if newMax > newMin {
		rout.SetDisplayMinimum(newMin)
		rout.SetDisplayMaximum(newMax)
	}
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
