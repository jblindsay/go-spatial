// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Feb. 2015.

package tools

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

type DifferenceFromMean struct {
	inputFile         string
	outputFile        string
	neighbourhoodSize int
	toolManager       *PluginToolManager
}

func (this *DifferenceFromMean) GetName() string {
	s := "DifferenceFromMean"
	return getFormattedToolName(s)
}

func (this *DifferenceFromMean) GetDescription() string {
	s := "Calculates the difference from mean"
	return getFormattedToolDescription(s)
}

func (this *DifferenceFromMean) GetHelpDocumentation() string {
	ret := "This tool is used to perform a fast difference from local mean filter operation."
	return ret
}

func (this *DifferenceFromMean) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *DifferenceFromMean) GetArgDescriptions() [][]string {
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

	ret[2][0] = "NeighbourhoodSize"
	ret[2][1] = "int"
	ret[2][2] = "The radius of the neighbourhood in grid cells"

	return ret
}

func (this *DifferenceFromMean) ParseArguments(args []string) {
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

	this.neighbourhoodSize = 1
	if len(strings.TrimSpace(args[2])) > 0 && args[2] != "not specified" {
		var err error
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(args[2]), 0, 0); err != nil {
			this.neighbourhoodSize = 1
			println(err)
		} else {
			this.neighbourhoodSize = int(val)
		}
	} else {
		this.neighbourhoodSize = 1
	}
	this.Run()
}

func (this *DifferenceFromMean) CollectArguments() {
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

	// get the neighbourhood radius argument
	print("Neighbourhood radius (grid cells): ")
	radiusStr, err := consolereader.ReadString('\n')
	if err != nil {
		this.neighbourhoodSize = 1
		println(err)
	}

	if len(strings.TrimSpace(radiusStr)) > 0 {
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(radiusStr), 0, 0); err != nil {
			this.neighbourhoodSize = 1
			println(err)
		} else {
			this.neighbourhoodSize = int(val)
		}
	} else {
		this.neighbourhoodSize = 1
	}

	this.Run()
}

func (this *DifferenceFromMean) Run() {
	start1 := time.Now()

	var progress, oldProgress, col, row int
	var z, sum float64
	var sumN, N int
	var x1, x2, y1, y2 int
	var outValue, mean float64

	println("Reading raster data...")
	rin, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		println(err.Error())
	}
	rows := rin.Rows
	columns := rin.Columns
	rowsLessOne := rows - 1
	nodata := rin.NoDataValue
	inConfig := rin.GetRasterConfig()
	minValue := rin.GetMinimumValue()
	// maxValue := rin.GetMaximumValue()
	// valueRange := maxValue - minValue
	k := minValue //+ valueRange/2.0
	multiplier := 100.0

	start2 := time.Now()

	// Initialize the arrays to hold the integral image, integral image squared, and number of valid cells
	I := make([][]uint64, rows)
	IN := make([][]int, rows)

	for row = 0; row < rows; row++ {
		I[row] = make([]uint64, columns)
		IN[row] = make([]int, columns)
	}

	// calculate the integral image
	printf("\rCalculating integral image (1 of 2): %v%%\n", 0)
	oldProgress = 0
	for row = 0; row < rows; row++ {
		sum = 0
		sumN = 0
		for col = 0; col < columns; col++ {
			z = rin.Value(row, col)
			if z == nodata {
				z = 0
			} else {
				z = (z - k) * multiplier
				sumN++
			}
			sum += z
			if row > 0 {
				I[row][col] = uint64(sum) + I[row-1][col]
				IN[row][col] = sumN + IN[row-1][col]
			} else {
				I[row][col] = uint64(sum)
				IN[row][col] = sumN
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress%5 == 0 && progress != oldProgress {
			printf("\rCalculating integral image (1 of 2): %v%%\n", progress)
			oldProgress = progress
		}
	}

	// output the data
	config := raster.NewDefaultRasterConfig()
	config.PreferredPalette = "blue_white_red.plt"
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

	printf("\rPerforming analysis (2 of 2): %v%%\n", 0)
	minVal := math.Inf(1)
	maxVal := math.Inf(-1)
	oldProgress = 0
	for row = 0; row < rows; row++ {
		y1 = row - this.neighbourhoodSize - 1
		if y1 < 0 {
			y1 = 0
		}
		if y1 >= rows {
			y1 = rows - 1
		}

		y2 = row + this.neighbourhoodSize
		if y2 < 0 {
			y2 = 0
		}
		if y2 >= rows {
			y2 = rows - 1
		}
		for col = 0; col < columns; col++ {
			z = rin.Value(row, col)
			if z != nodata {
				x1 = col - this.neighbourhoodSize - 1
				if x1 < 0 {
					x1 = 0
				}
				if x1 >= columns {
					x1 = columns - 1
				}

				x2 = col + this.neighbourhoodSize
				if x2 < 0 {
					x2 = 0
				}
				if x2 >= columns {
					x2 = columns - 1
				}

				N = IN[y2][x2] + IN[y1][x1] - IN[y1][x2] - IN[y2][x1]
				if N > 0 {
					sum = float64(I[y2][x2]+I[y1][x1]-I[y1][x2]-I[y2][x1]) / multiplier
					mean = sum / float64(N)
					outValue = (z - k) - mean
					rout.SetValue(row, col, outValue)
					if outValue > maxVal {
						maxVal = outValue
					}
					if outValue < minVal {
						minVal = outValue
					}
				} else {
					rout.SetValue(row, col, 0)
				}
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress%5 == 0 && progress != oldProgress {
			printf("\rPerforming analysis (2 of 2): %v%%\n", progress)
			oldProgress = progress
		}
	}

	elapsed := time.Since(start2)
	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by DeviationFromMean tool"))
	rout.AddMetadataEntry(fmt.Sprintf("Window size: %v", (this.neighbourhoodSize*2 + 1)))
	palVal := math.Min(math.Abs(minVal), maxVal)
	config.DisplayMinimum = -palVal
	config.DisplayMaximum = palVal
	rout.SetRasterConfig(config)
	rout.Save()

	println("Operation complete!")

	value := fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	println(value)

	overallTime := time.Since(start1)
	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}
