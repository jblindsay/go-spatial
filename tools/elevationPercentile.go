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

type ElevationPercentile struct {
	inputFile         string
	outputFile        string
	neighbourhoodSize int
	numBins           uint32
	toolManager       *PluginToolManager
}

func (this *ElevationPercentile) GetName() string {
	s := "ElevationPercentile"
	return getFormattedToolName(s)
}

func (this *ElevationPercentile) GetDescription() string {
	s := "Calculates the local elevation percentile for a DEM"
	return getFormattedToolDescription(s)
}

func (this *ElevationPercentile) GetHelpDocumentation() string {
	ret := "This tool is used to remove the sinks (i.e. topographic depressions and flat areas) from digital elevation models (DEMs) using an efficient depression filling method. Note that the BreachDepressions tool is the preferred method of creating a depressionless DEM."
	return ret
}

func (this *ElevationPercentile) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *ElevationPercentile) GetArgDescriptions() [][]string {
	numArgs := 4

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

	ret[3][0] = "NumBins"
	ret[3][1] = "int"
	ret[3][2] = "The number of bins used to calculate the histogram"

	return ret
}

func (this *ElevationPercentile) ParseArguments(args []string) {
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
			println(err)
		} else {
			this.neighbourhoodSize = int(val)
		}
	}

	this.numBins = 1
	if len(strings.TrimSpace(args[3])) > 0 && args[3] != "not specified" {
		var err error
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(args[3]), 0, 0); err != nil {
			println(err)
		} else {
			this.numBins = uint32(val)
		}
	}

	this.Run()
}

func (this *ElevationPercentile) CollectArguments() {
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
	this.neighbourhoodSize = 1
	radiusStr, err := consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}

	if len(strings.TrimSpace(radiusStr)) > 0 {
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(radiusStr), 0, 0); err != nil {
			println(err)
		} else {
			this.neighbourhoodSize = int(val)
		}
	}

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
			this.numBins = uint32(val)
		}
	}

	this.Run()
}

func (this *ElevationPercentile) Run() {
	start1 := time.Now()

	var progress, oldProgress, col, row int
	var i, j, bin, highResNumBins uint32
	var z, percentile float64
	var N, numLess, binRunningTotal uint32
	var x1, x2, y1, y2 int
	var a, b, c, d, e, f, g, rowSum []uint32

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

	highResNumBins = 10000
	highResBinSize := valueRange / float64(highResNumBins)

	primaryHisto := make([]uint32, highResNumBins)
	var numValidCells uint32 = 0
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			z = rin.Value(row, col)
			if z != nodata {
				i = uint32(math.Floor((z - minValue) / highResBinSize))
				//				if i == this.numBins {
				//					i = this.numBins - 1
				//				}
				if i >= highResNumBins {
					i = highResNumBins - 1
				}
				primaryHisto[i]++
				numValidCells++
			}
		}
	}
	quantileProportion := numValidCells / this.numBins
	binNumMap := make([]uint32, highResNumBins)
	binTotal := make([]uint32, this.numBins)
	valProbMap := make([]float64, highResNumBins)
	binRunningTotal = 0
	bin = 0
	for i = 0; i < highResNumBins; i++ {
		binRunningTotal += primaryHisto[i]
		if binRunningTotal > quantileProportion {
			if bin < this.numBins-1 {
				bin++
				binRunningTotal = primaryHisto[i]
			}
		}
		binNumMap[i] = bin
		binTotal[bin] += primaryHisto[i]
		valProbMap[i] = float64(binRunningTotal)
		//primaryHisto[i] += primaryHisto[i-1]
	}

	for i = 0; i < highResNumBins; i++ {
		valProbMap[i] = valProbMap[i] / float64(binTotal[binNumMap[i]])
	}

	//	for i = 0; i < highResNumBins; i++ {
	//		primaryHisto[i] = uint32(math.Floor(cdf[i] / quantileProportion))
	//		if primaryHisto[i] == this.numBins {
	//			primaryHisto[i] = this.numBins - 1
	//		}
	//	}

	//	binLowerValue := make([]float64, this.numBins)
	//	binSize := make([]float64, this.numBins)
	//	for i = 0; i < this.numBins; i++ {
	//		binLowerValue[i] = minValue + float64(i)*binSize
	//	}
	//	bin = -1
	//	for i = 0; i < highResNumBins; i++ {
	//		if primaryHisto[i] > bin {
	//			bin = primaryHisto[i]
	//			// what elevation does this bin correpsond to?
	//			binLowerValue[bin] = minValue + float64(i)*highResBinSize
	//			if bin > 0 {
	//				binSize[bin-1] = (minValue + float64(i-1)*highResBinSize) - binLowerValue[bin-1]
	//			}
	//		}
	//	}
	//	binSize[this.numBins-1] = maxValue - binLowerValue[this.numBins-1]

	//	for i = 0; i < this.numBins; i++ {
	//		println(binLowerValue[i], binSize[i])
	//	}

	histoImage := make([][][]uint32, rows)

	oldProgress = -1
	for row = 0; row < rows; row++ {
		histoImage[row] = make([][]uint32, columns)
		rowSum = make([]uint32, this.numBins)
		for col = 0; col < columns; col++ {
			z = rin.Value(row, col)
			if z != nodata {
				//bin = int(math.Floor((z - minValue) / binSize))
				i = uint32(math.Floor((z - minValue) / highResBinSize))
				if i >= highResNumBins {
					i = highResNumBins - 1
				}
				bin = binNumMap[i]
				rowSum[bin]++
			}
			histoImage[row][col] = make([]uint32, this.numBins)
			if row > 0 {
				for i = 0; i < this.numBins; i++ {
					histoImage[row][col][i] = rowSum[i] + histoImage[row-1][col][i]
				}
			} else {
				for i = 0; i < this.numBins; i++ {
					histoImage[row][col][i] = rowSum[i]
				}
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress%5 == 0 && progress != oldProgress {
			printf("Calculating integral histogram (1 of 2): %v%%\n", progress)
			oldProgress = progress
		}
	}

	// create the output raster
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

	e = make([]uint32, this.numBins)
	f = make([]uint32, this.numBins)
	g = make([]uint32, this.numBins)

	oldProgress = -1
	for row = 0; row < rows; row++ {
		y1 = row - this.neighbourhoodSize
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
				//bin = int(math.Floor((z - minValue) / binSize))
				j = uint32(math.Floor((z - minValue) / highResBinSize))
				if j >= highResNumBins {
					j = highResNumBins - 1
				}
				bin = binNumMap[j]

				x1 = col - this.neighbourhoodSize
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

				a = histoImage[y2][x2]
				b = histoImage[y1][x1]
				c = histoImage[y1][x2]
				d = histoImage[y2][x1]

				for i = 0; i < this.numBins; i++ {
					e[i] = a[i] + b[i]
				}
				for i = 0; i < this.numBins; i++ {
					f[i] = e[i] - c[i]
				}
				for i = 0; i < this.numBins; i++ {
					g[i] = f[i] - d[i]
				}

				N = 0
				numLess = 0
				for i = 0; i < this.numBins; i++ {
					N += g[i]
					if i < bin {
						numLess += g[i]
					}
				}

				if N > 0 {
					//percentile = 100.0 * float64(g[bin]) / float64(N) // only used for accuracy assessment
					percentile = 100.0 * (float64(numLess) + valProbMap[j]*float64(g[bin])) / float64(N)
					//percentile = 100.0 * (float64(numLess) + (z-binLowerValue[bin])/binSize[bin]*float64(g[bin])) / float64(N)
					rout.SetValue(row, col, percentile)
				}
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress%5 == 0 && progress != oldProgress {
			printf("Performing analysis (2 of 2): %v%%\n", progress)
			oldProgress = progress
		}
	}

	println("Saving data...")

	elapsed := time.Since(start2)
	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by ElevationPercentile tool"))
	rout.AddMetadataEntry(fmt.Sprintf("Window size: %v", (this.neighbourhoodSize*2 + 1)))
	rout.AddMetadataEntry(fmt.Sprintf("Num. histogram bins: %v", this.numBins))
	config.DisplayMinimum = 0
	config.DisplayMaximum = 100
	rout.SetRasterConfig(config)
	rout.Save()

	println("Operation complete!")

	value := fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	println(value)

	overallTime := time.Since(start1)
	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}
