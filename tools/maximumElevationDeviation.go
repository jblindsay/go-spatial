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
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

type MaximumElevationDeviation struct {
	inputFile         string
	magOutputFile     string
	scaleOutputFile   string
	minNeighbourhood  int
	maxNeighbourhood  int
	neighbourhoodStep int
	toolManager       *PluginToolManager
}

func (this *MaximumElevationDeviation) GetName() string {
	s := "MaxElevationDeviation"
	return getFormattedToolName(s)
}

func (this *MaximumElevationDeviation) GetDescription() string {
	s := "Calculates the maximum elevation deviation across a range of scales"
	return getFormattedToolDescription(s)
}

func (this *MaximumElevationDeviation) GetHelpDocumentation() string {
	ret := "This tool is used to remove the sinks (i.e. topographic depressions and flat areas) from digital elevation models (DEMs) using an efficient depression filling method. Note that the BreachDepressions tool is the preferred method of creating a depressionless DEM."
	return ret
}

func (this *MaximumElevationDeviation) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *MaximumElevationDeviation) GetArgDescriptions() [][]string {
	numArgs := 6

	ret := make([][]string, numArgs)
	for i := range ret {
		ret[i] = make([]string, 3)
	}
	ret[0][0] = "InputDEM"
	ret[0][1] = "string"
	ret[0][2] = "The input DEM name, with directory and file extension"

	ret[1][0] = "OutputMagnitudeFile"
	ret[1][1] = "string"
	ret[1][2] = "The magnitude output filename, with directory and file extension"

	ret[2][0] = "OutputScaleFile"
	ret[2][1] = "string"
	ret[2][2] = "The scale output filename, with directory and file extension"

	ret[3][0] = "MinNeighbourhoodSize"
	ret[3][1] = "int"
	ret[3][2] = "The starting radius of the neighbourhood in grid cells"

	ret[4][0] = "MaxNeighbourhoodSize"
	ret[4][1] = "int"
	ret[4][2] = "The ending radius of the neighbourhood in grid cells"

	ret[5][0] = "NeighbourhoodStep"
	ret[5][1] = "int"
	ret[5][2] = "The neighbourhood step size in grid cells"

	return ret
}

func (this *MaximumElevationDeviation) ParseArguments(args []string) {
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
	this.magOutputFile = outputFile

	outputFile = args[2]
	outputFile = strings.TrimSpace(outputFile)
	if !strings.Contains(outputFile, pathSep) {
		outputFile = this.toolManager.workingDirectory + outputFile
	}
	rasterType, err = raster.DetermineRasterFormat(outputFile)
	if rasterType == raster.RT_UnknownRaster || err == raster.UnsupportedRasterFormatError {
		outputFile = outputFile + ".tif" // default to a geotiff
	}
	this.scaleOutputFile = outputFile

	this.minNeighbourhood = 1
	if len(strings.TrimSpace(args[3])) > 0 && args[3] != "not specified" {
		var err error
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(args[3]), 0, 0); err != nil {
			println(err)
		} else {
			this.minNeighbourhood = int(val)
		}
	}

	this.maxNeighbourhood = 3
	if len(strings.TrimSpace(args[4])) > 0 && args[4] != "not specified" {
		var err error
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(args[4]), 0, 0); err != nil {
			println(err)
		} else {
			this.maxNeighbourhood = int(val)
		}
	}

	this.neighbourhoodStep = 1
	if len(strings.TrimSpace(args[5])) > 0 && args[5] != "not specified" {
		var err error
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(args[5]), 0, 0); err != nil {
			println(err)
		} else {
			this.neighbourhoodStep = int(val)
		}
	}

	this.Run()
}

func (this *MaximumElevationDeviation) CollectArguments() {
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
	print("Enter the magnitude output file name (incl. file extension): ")
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
	this.magOutputFile = outputFile

	// get the output file name
	print("Enter the scale output file name (incl. file extension): ")
	outputFile, err = consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}
	outputFile = strings.TrimSpace(outputFile)
	if !strings.Contains(outputFile, pathSep) {
		outputFile = this.toolManager.workingDirectory + outputFile
	}
	rasterType, err = raster.DetermineRasterFormat(outputFile)
	if rasterType == raster.RT_UnknownRaster || err == raster.UnsupportedRasterFormatError {
		outputFile = outputFile + ".tif" // default to a geotiff
	}
	this.scaleOutputFile = outputFile

	// get the min neighbourhood radius argument
	print("Min. neighbourhood radius (grid cells): ")
	radiusStr, err := consolereader.ReadString('\n')
	if err != nil {
		this.minNeighbourhood = 1
		println(err)
	}

	if len(strings.TrimSpace(radiusStr)) > 0 {
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(radiusStr), 0, 0); err != nil {
			this.minNeighbourhood = 1
			println(err)
		} else {
			this.minNeighbourhood = int(val)
		}
	} else {
		this.minNeighbourhood = 1
	}

	// get the max neighbourhood radius argument
	print("Max. neighbourhood radius (grid cells): ")
	radiusStr, err = consolereader.ReadString('\n')
	if err != nil {
		this.maxNeighbourhood = 3
		println(err)
	}

	if len(strings.TrimSpace(radiusStr)) > 0 {
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(radiusStr), 0, 0); err != nil {
			this.maxNeighbourhood = 3
			println(err)
		} else {
			this.maxNeighbourhood = int(val)
		}
	} else {
		this.maxNeighbourhood = 3
	}

	// get the neighbourhood step argument
	print("Neighbourhood step size (grid cells): ")
	this.neighbourhoodStep = 1
	radiusStr, err = consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}

	if len(strings.TrimSpace(radiusStr)) > 0 {
		var val int64
		if val, err = strconv.ParseInt(strings.TrimSpace(radiusStr), 0, 0); err != nil {
			println(err)
		} else {
			this.neighbourhoodStep = int(val)
		}
	}

	this.Run()
}

func (this *MaximumElevationDeviation) Run() {
	start1 := time.Now()

	var progress, oldProgress, col, row int
	var z, sum, sumSqr float64
	var sumN int //, N int
	// var x1, x2, y1, y2 int
	// var outValue, v, s, m float64
	var str string

	fmt.Println("Reading raster data...")
	rin, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		fmt.Println(err.Error())
	}
	rows := rin.Rows
	columns := rin.Columns
	rowsLessOne := rows - 1
	nodata := rin.NoDataValue
	inConfig := rin.GetRasterConfig()
	minValue := rin.GetMinimumValue()
	maxValue := rin.GetMaximumValue()
	valueRange := maxValue - minValue
	k := minValue + valueRange/2.0

	start2 := time.Now()

	I := make([][]float64, rows)
	I2 := make([][]float64, rows)
	IN := make([][]int, rows)
	maxVal := make([][]float64, rows)
	scaleVal := make([][]int, rows)
	zVal := make([][]float64, rows)

	for row = 0; row < rows; row++ {
		I[row] = make([]float64, columns)
		I2[row] = make([]float64, columns)
		IN[row] = make([]int, columns)
		maxVal[row] = make([]float64, columns)
		scaleVal[row] = make([]int, columns)
		zVal[row] = make([]float64, columns)
	}

	// calculate the integral image
	oldProgress = -1
	for row = 0; row < rows; row++ {
		sum = 0
		sumSqr = 0
		sumN = 0
		for col = 0; col < columns; col++ {
			z = rin.Value(row, col)
			zVal[row][col] = z
			if z == nodata {
				z = 0
			} else {
				z = z - k
				sumN++
			}
			sum += z
			sumSqr += z * z
			if row > 0 {
				I[row][col] = sum + I[row-1][col]
				I2[row][col] = sumSqr + I2[row-1][col]
				IN[row][col] = sumN + IN[row-1][col]
			} else {
				I[row][col] = sum
				I2[row][col] = sumSqr
				IN[row][col] = sumN
			}
			maxVal[row][col] = -math.MaxFloat32

		}
		progress = int(100.0 * row / rowsLessOne)
		if progress%5 == 0 && progress != oldProgress {
			fmt.Printf("Calculating integral image: %v%%\n", progress)
			oldProgress = progress
		}
	}

	// fmt.Println("\r                                    ")

	numCPUs := runtime.NumCPU()

	oldProgress = -1
	loopNum := 1
	numLoops := int((this.maxNeighbourhood-this.minNeighbourhood)/this.neighbourhoodStep) + 1
	for neighbourhood := this.minNeighbourhood; neighbourhood <= this.maxNeighbourhood; neighbourhood += this.neighbourhoodStep {
		c1 := make(chan bool)
		runtime.GOMAXPROCS(numCPUs)
		var wg sync.WaitGroup
		startingRow := 0
		var rowBlockSize int = rows / numCPUs

		for startingRow < rows {
			endingRow := startingRow + rowBlockSize
			if endingRow >= rows {
				endingRow = rows - 1
			}
			wg.Add(1)
			go func(rowSt, rowEnd int) {
				defer wg.Done()
				var x1, x2, y1, y2, N int
				var outValue, z, sum, mean float64
				var v, s float64
				for row := rowSt; row <= rowEnd; row++ {
					y1 = row - neighbourhood - 1
					if y1 < 0 {
						y1 = 0
					}
					if y1 >= rows {
						y1 = rows - 1
					}

					y2 = row + neighbourhood
					if y2 < 0 {
						y2 = 0
					}
					if y2 >= rows {
						y2 = rows - 1
					}
					// floatData := make([]float64, columns)
					for col := 0; col < columns; col++ {
						z = rin.Value(row, col)
						if z != nodata {
							x1 = col - neighbourhood - 1
							if x1 < 0 {
								x1 = 0
							}
							if x1 >= columns {
								x1 = columns - 1
							}

							x2 = col + neighbourhood
							if x2 < 0 {
								x2 = 0
							}
							if x2 >= columns {
								x2 = columns - 1
							}
							N = IN[y2][x2] + IN[y1][x1] - IN[y1][x2] - IN[y2][x1]
							if N > 0 {
								sum = I[y2][x2] + I[y1][x1] - I[y1][x2] - I[y2][x1]
								sumSqr = I2[y2][x2] + I2[y1][x1] - I2[y1][x2] - I2[y2][x1]
								v = (sumSqr - (sum*sum)/float64(N)) / float64(N)
								if v > 0 {
									s = math.Sqrt(v)
									mean = sum / float64(N)
									outValue = ((z - k) - mean) / s
									if math.Abs(outValue) > maxVal[row][col] {
										maxVal[row][col] = math.Abs(outValue)
										if outValue >= 0 {
											scaleVal[row][col] = neighbourhood
										} else {
											scaleVal[row][col] = -neighbourhood
										}
									}
								}
							}

							// N = IN[y2][x2] + IN[y1][x1] - IN[y1][x2] - IN[y2][x1]
							// if N > 0 {
							// 	sum = I[y2][x2] + I[y1][x1] - I[y1][x2] - I[y2][x1]
							// 	sumSqr = I2[y2][x2] + I2[y1][x1] - I2[y1][x2] - I2[y2][x1]
							// 	v = (sumSqr - (sum*sum)/float64(N)) / float64(N)
							// 	if v > 0 {
							// 		s = math.Sqrt(v)
							// 		mean = sum / float64(N)
							// 		outValue = ((z - k) - mean) / s
							// 		floatData[col] = outValue
							// 	} else {
							// 		floatData[col] = 0
							// 	}
							// } else {
							// 	floatData[col] = 0.0
							// }

						} // else {
						//	floatData[col] = nodata
						//}
					}
					//rout.SetRowValues(row, floatData)
					c1 <- true // row completed
				}

			}(startingRow, endingRow)
			startingRow = endingRow + 1
		}

		oldProgress = 0
		for rowsCompleted := 0; rowsCompleted < rows; rowsCompleted++ {
			<-c1 // a row has successfully completed
			progress = int(100.0 * float64(rowsCompleted) / float64(rowsLessOne))
			if progress != oldProgress {
				str = fmt.Sprintf("Loop %v of %v", loopNum, numLoops)
				fmt.Printf("%s: %v%%\n", str, progress)

				// fmt.Printf("Progress: %v%%\n", progress)
				oldProgress = progress
			}
		}

		wg.Wait()

		// for row = 0; row < rows; row++ {
		// 	y1 = row - neighbourhood - 1
		// 	if y1 < 0 {
		// 		y1 = 0
		// 	}
		// 	if y1 >= rows {
		// 		y1 = rows - 1
		// 	}
		//
		// 	y2 = row + neighbourhood
		// 	if y2 < 0 {
		// 		y2 = 0
		// 	}
		// 	if y2 >= rows {
		// 		y2 = rows - 1
		// 	}
		//
		// 	for col = 0; col < columns; col++ {
		// 		z = zVal[row][col]
		// 		if z != nodata {
		// 			x1 = col - neighbourhood - 1
		// 			if x1 < 0 {
		// 				x1 = 0
		// 			}
		// 			if x1 >= columns {
		// 				x1 = columns - 1
		// 			}
		//
		// 			x2 = col + neighbourhood
		// 			if x2 < 0 {
		// 				x2 = 0
		// 			}
		// 			if x2 >= columns {
		// 				x2 = columns - 1
		// 			}
		//
		// 			N = IN[y2][x2] + IN[y1][x1] - IN[y1][x2] - IN[y2][x1]
		// 			if N > 0 {
		// 				sum = I[y2][x2] + I[y1][x1] - I[y1][x2] - I[y2][x1]
		// 				sumSqr = I2[y2][x2] + I2[y1][x1] - I2[y1][x2] - I2[y2][x1]
		// 				v = (sumSqr - (sum*sum)/float64(N)) / float64(N)
		// 				if v > 0 {
		// 					s = math.Sqrt(v)
		// 					m = sum / float64(N)
		// 					outValue = ((z - k) - m) / s
		// 					if math.Abs(outValue) > maxVal[row][col] {
		// 						maxVal[row][col] = math.Abs(outValue)
		// 						if outValue >= 0 {
		// 							//output.setValue(row, col, neighbourhood)
		// 							scaleVal[row][col] = neighbourhood
		// 						} else {
		// 							//output.setValue(row, col, -neighbourhood)
		// 							scaleVal[row][col] = -neighbourhood
		// 						}
		// 						//output2.setValue(row, col, outValue)
		// 					}
		// 				}
		// 			}
		// 		}
		// 	}
		// 	progress = int(100.0 * row / rowsLessOne)
		// 	if progress != oldProgress {
		// 		str = fmt.Sprintf("Loop %v of %v", loopNum, numLoops)
		// 		printf("\r%s: %v%%", str, progress)
		// 		oldProgress = progress
		// 	}
		// }

		loopNum++
	}

	// output the data
	config := raster.NewDefaultRasterConfig()
	config.PreferredPalette = "blue_white_red.plt"
	config.DataType = raster.DT_FLOAT32
	config.NoDataValue = nodata
	config.InitialValue = nodata
	config.CoordinateRefSystemWKT = inConfig.CoordinateRefSystemWKT
	config.EPSGCode = inConfig.EPSGCode
	rout1, err := raster.CreateNewRaster(this.magOutputFile, rows, columns,
		rin.North, rin.South, rin.East, rin.West, config)
	if err != nil {
		println("Failed to write raster")
		return
	}

	config2 := raster.NewDefaultRasterConfig()
	config2.PreferredPalette = "blue_white_red.plt"
	config2.DataType = raster.DT_FLOAT32
	config2.NoDataValue = nodata
	config2.InitialValue = nodata
	config2.CoordinateRefSystemWKT = inConfig.CoordinateRefSystemWKT
	config2.EPSGCode = inConfig.EPSGCode
	rout2, err := raster.CreateNewRaster(this.scaleOutputFile, rows, columns,
		rin.North, rin.South, rin.East, rin.West, config2)
	if err != nil {
		println("Failed to write raster")
		return
	}

	config.DisplayMinimum = -3.0
	config.DisplayMaximum = 3.0

	config2.PreferredPalette = "imhof1.plt"
	rout2.SetRasterConfig(config2)

	println("Saving the outputs...")
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			if maxVal[row][col] > -math.MaxFloat32 {
				if scaleVal[row][col] >= 0 {
					rout1.SetValue(row, col, maxVal[row][col])
					rout2.SetValue(row, col, float64(scaleVal[row][col]))
				} else {
					rout1.SetValue(row, col, -maxVal[row][col])
					rout2.SetValue(row, col, float64(-scaleVal[row][col]))
				}
			}
		}
	}

	rout1.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)
	rout1.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout1.AddMetadataEntry(fmt.Sprintf("Created by ElevationPercentile tool"))
	rout1.AddMetadataEntry(fmt.Sprintf("Min. window size: %v", (this.minNeighbourhood*2 + 1)))
	rout1.AddMetadataEntry(fmt.Sprintf("Max. window size: %v", (this.maxNeighbourhood*2 + 1)))
	rout1.AddMetadataEntry(fmt.Sprintf("Step size: %v", this.neighbourhoodStep))

	rout2.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	rout2.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout2.AddMetadataEntry(fmt.Sprintf("Created by ElevationPercentile tool"))
	rout2.AddMetadataEntry(fmt.Sprintf("Min. window size: %v", (this.minNeighbourhood*2 + 1)))
	rout2.AddMetadataEntry(fmt.Sprintf("Max. window size: %v", (this.maxNeighbourhood*2 + 1)))
	rout2.AddMetadataEntry(fmt.Sprintf("Step size: %v", this.neighbourhoodStep))

	overallTime := time.Since(start1)
	rout1.SetRasterConfig(config)
	rout1.Save()
	rout2.SetRasterConfig(config2)
	rout2.Save()

	println("Operation complete!")

	value := fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	println(value)

	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}
