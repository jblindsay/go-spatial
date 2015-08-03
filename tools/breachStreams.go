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
	"github.com/jblindsay/go-spatial/structures"
)

type BreachStreams struct {
	streamFile  string
	demFile     string
	outputFile  string
	toolManager *PluginToolManager
}

func (this *BreachStreams) GetName() string {
	s := "BreachStreams"
	return getFormattedToolName(s)
}

func (this *BreachStreams) GetDescription() string {
	s := "Breaches a stream network into a DEM"
	return getFormattedToolDescription(s)
}

func (this *BreachStreams) GetHelpDocumentation() string {
	ret := "This tool is used to remove the sinks (i.e. topographic depressions and flat areas) from digital elevation models (DEMs) using a highly efficient and flexible breaching, or carving, method."
	return ret
}

func (this *BreachStreams) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

// Can be called to gather a listing of the arguments required to run this tool.
func (this *BreachStreams) GetArgDescriptions() [][]string {
	numArgs := 3
	ret := structures.Create2dStringArray(numArgs, 3)

	ret[0][0] = "InputStream"
	ret[0][1] = "string"
	ret[0][2] = "The input stream raster file name with file extension"

	ret[1][0] = "InputDEM"
	ret[1][1] = "string"
	ret[1][2] = "The input DEM name with file extension"

	ret[2][0] = "OutputFile"
	ret[2][1] = "string"
	ret[2][2] = "The output filename with file extension"

	return ret
}

// ParseArguments is used when the tool is run using command-line args
// rather than in interactive input/output mode.
func (this *BreachStreams) ParseArguments(args []string) {
	streamFile := args[0]
	streamFile = strings.TrimSpace(streamFile)
	if !strings.Contains(streamFile, pathSep) {
		streamFile = this.toolManager.workingDirectory + streamFile
	}
	this.streamFile = streamFile
	// see if the file exists
	if _, err := os.Stat(this.streamFile); os.IsNotExist(err) {
		printf("no such file or directory: %s\n", this.streamFile)
		return
	}

	demFile := args[1]
	demFile = strings.TrimSpace(demFile)
	if !strings.Contains(demFile, pathSep) {
		demFile = this.toolManager.workingDirectory + demFile
	}
	this.demFile = demFile
	// see if the file exists
	if _, err := os.Stat(this.demFile); os.IsNotExist(err) {
		printf("no such file or directory: %s\n", this.demFile)
		return
	}

	outputFile := args[2]
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

func (this *BreachStreams) CollectArguments() {
	consolereader := bufio.NewReader(os.Stdin)

	// get the input streams file name
	print("Enter the streams raster file name (incl. file extension): ")
	streamFile, err := consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}
	streamFile = strings.TrimSpace(streamFile)
	if !strings.Contains(streamFile, pathSep) {
		streamFile = this.toolManager.workingDirectory + streamFile
	}
	this.streamFile = streamFile
	// see if the file exists
	if _, err := os.Stat(this.streamFile); os.IsNotExist(err) {
		printf("no such file or directory: %s\n", this.streamFile)
		return
	}

	// get the input DEM file name
	print("Enter the DEM file name (incl. file extension): ")
	demFile, err := consolereader.ReadString('\n')
	if err != nil {
		println(err)
	}
	demFile = strings.TrimSpace(demFile)
	if !strings.Contains(demFile, pathSep) {
		demFile = this.toolManager.workingDirectory + demFile
	}
	this.demFile = demFile
	// see if the file exists
	if _, err := os.Stat(this.demFile); os.IsNotExist(err) {
		printf("no such file or directory: %s\n", this.demFile)
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

func (this *BreachStreams) Run() {
	start1 := time.Now()

	var progress, oldProgress, col, row, i, n int
	var colN, rowN, r, c, flatindex int
	numSolvedCells := 0
	var dir byte
	var z, zN, lowestNeighbour, s, sN float64
	var zTest, zN2, zN3 float64
	var gc gridCell
	var p int64
	var isPit, isEdgeCell, isStream bool
	numPits := 0
	numPitsSolved := 0
	numUnsolvedPits := 0
	numValidCells := 0
	var isActive bool
	dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
	dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
	backLink := [8]byte{5, 6, 7, 8, 1, 2, 3, 4}

	println("Reading input data...")
	dem, err := raster.CreateRasterFromFile(this.demFile)
	if err != nil {
		println(err.Error())
	}
	demConfig := dem.GetRasterConfig()
	rows := dem.Rows
	columns := dem.Columns
	rowsLessOne := rows - 1
	numCellsTotal := rows * columns
	nodata := dem.NoDataValue
	paletteName := demConfig.PreferredPalette
	minVal := dem.GetMinimumValue()
	elevDigits := len(strconv.Itoa(int(dem.GetMaximumValue() - minVal)))
	elevMultiplier := math.Pow(10, float64(8-elevDigits))
	SMALL_NUM := 1 / elevMultiplier * 10
	POS_INF := math.Inf(1)

	streams, err := raster.CreateRasterFromFile(this.streamFile)
	if err != nil {
		println(err.Error())
	}
	if streams.Rows != rows || streams.Columns != columns {
		println("The input rasters must be of the same dimensions.")
		return
	}
	streamsNodata := streams.NoDataValue

	start2 := time.Now()

	output := make([][]float64, rows+2)
	pits := make([][]bool, rows+2)
	inQueue := make([][]bool, rows+2)
	flowdir := make([][]byte, rows+2)

	for i = 0; i < rows+2; i++ {
		output[i] = make([]float64, columns+2)
		pits[i] = make([]bool, columns+2)
		inQueue[i] = make([]bool, columns+2)
		flowdir[i] = make([]byte, columns+2)
	}

	pq := NewPQueue()

	//	oldProgress = 0
	//	for row = 0; row < rows; row++ {
	//		for col = 0; col < columns; col++ {
	//			z = dem.Value(row, col)
	//			output[row+1][col+1] = z
	//			flowdir[row+1][col+1] = 0
	//			if z != nodata {
	//				s = streams.Value(row, col)
	//				if s != streamsNodata && s > 0 {
	//					lowestNeighbour = POS_INF
	//					for n = 0; n < 8; n++ {
	//						sN = streams.Value(row+dY[n], col+dX[n])
	//						if sN != streamsNodata && sN > 0 {
	//							zN = dem.Value(row+dY[n], col+dX[n])
	//							if zN < lowestNeighbour {
	//								lowestNeighbour = zN
	//							}
	//						}
	//					}
	//					if lowestNeighbour < z {
	//						output[row+1][col+1] = lowestNeighbour - SMALL_NUM
	//					}
	//				}
	//			}
	//		}
	//		progress = int(100.0 * row / rowsLessOne)
	//		if progress != oldProgress {
	//			printf("\rBreaching DEM (1 of 3): %v%%", progress)
	//			oldProgress = progress
	//		}
	//	}

	// find the pit cells and initialize the grids
	printf("\rBreaching DEM (1 of 2): %v%%", 0)
	oldProgress = 0
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			z = dem.Value(row, col)
			output[row+1][col+1] = z
			flowdir[row+1][col+1] = 0
			//z = output[row+1][col+1]
			if z != nodata {
				isPit = true
				isEdgeCell = false
				lowestNeighbour = POS_INF
				s = streams.Value(row, col)
				if s != streamsNodata && s > 0 {
					isStream = true
				} else {
					isStream = false
				}

				for n = 0; n < 8; n++ {
					zN = dem.Value(row+dY[n], col+dX[n])
					//zN = output[row+dY[n]+1][col+dX[n]+1]
					sN = streams.Value(row+dY[n], col+dX[n])
					if zN != nodata && zN < z { // there's a lower cell
						if !isStream {
							isPit = false
							//break
						} else {
							if sN != streamsNodata && sN > 0 { // there's a lower stream cell; it's not a stream pit
								isPit = false
								//break
							}
						}

					} else if zN == nodata {
						isEdgeCell = true
					} else {
						if zN < lowestNeighbour {
							lowestNeighbour = zN
						}
					}
				}

				if isEdgeCell {
					gc = newGridCell(row+1, col+1, 0)
					if isStream {
						p = int64(int64(z*elevMultiplier) * 10000)
						// given their much higher priorities, stream cells will always
						// be visited before non-stream cells when they are present
						// in the queue.
					} else {
						p = int64(10000000000000 + int64(z*elevMultiplier)*10000)
					}
					pq.Push(gc, p)
					inQueue[row+1][col+1] = true
				}
				if isPit {
					if !isEdgeCell {
						pits[row+1][col+1] = true
						numPits++
					}
					/* raising a pit cell to just lower than the
					 *  elevation of its lowest neighbour will
					 *  reduce the length and depth of the trench
					 *  that is necessary to eliminate the pit
					 *  by quite a bit on average.
					 */
					if lowestNeighbour != POS_INF && !isStream { // this shouldn't be done for stream cells
						output[row+1][col+1] = lowestNeighbour - SMALL_NUM
					}
					//}
				}
				numValidCells++
			} else {
				numSolvedCells++
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress != oldProgress {
			printf("\rBreaching DEM (1 of 2): %v%%", progress)
			oldProgress = progress
		}
	}

	for row = 0; row < rows+2; row++ {
		output[row][0] = nodata
		output[row][columns+1] = nodata
		flowdir[row][0] = 0
		flowdir[row][columns+1] = 0
	}

	for col = 0; col < columns+2; col++ {
		output[0][col] = nodata
		output[rows+1][col] = nodata
		flowdir[0][col] = 0
		flowdir[rows+1][col] = 0
	}

	// now breach
	printf("\r                                                                 ")
	oldProgress = int(100.0 * numSolvedCells / numCellsTotal)
	printf("\rBreaching DEM (2 of 2): %v%%", oldProgress)

	// Perform a complete breaching solution; there will be no subseqent filling
	for numPitsSolved < numPits {
		gc = pq.Pop()
		row = gc.row
		col = gc.column
		flatindex = gc.flatIndex

		//		s = streams.Value(row, col)
		//		if s != streamsNodata && s > 0 {
		//			output[row+1][col+1] -= 10.0
		//		}

		for i = 0; i < 8; i++ {
			rowN = row + dY[i]
			colN = col + dX[i]
			zN = output[rowN][colN]
			if zN != nodata && !inQueue[rowN][colN] {
				flowdir[rowN][colN] = backLink[i]
				if pits[rowN][colN] {
					numPitsSolved++
					// trace the flowpath back until you find a lower cell
					zTest = zN
					r = rowN
					c = colN
					isActive = true
					for isActive {
						zTest -= SMALL_NUM // ensures a small increment slope
						s = streams.Value(r, c)
						if s > 0 && s != streamsNodata {
							// is there a neighbouring non-stream cell that is lower than zTest?
							lowestNeighbour = POS_INF // this will actually be the lowest non-stream neighbour
							for n = 0; n < 8; n++ {
								sN = streams.Value(r+dY[n], c+dX[n])
								zN3 = output[r+dY[n]][c+dX[n]]
								if (sN == 0 || sN == streamsNodata) && zN3 != nodata { // it's a non-stream but valid neighbour
									if zN3 < lowestNeighbour {
										lowestNeighbour = zN3
									}
								}
							}
							if lowestNeighbour < zTest {
								zTest = lowestNeighbour - SMALL_NUM
							}
						}
						dir = flowdir[r][c]
						if dir > 0 {
							r += dY[dir-1]
							c += dX[dir-1]
							zN2 = output[r][c]
							if zN2 <= zTest || zN2 == nodata {
								// a lower grid cell or edge has been found
								isActive = false
							} else {
								output[r][c] = zTest
							}
						} else {
							// a pit has been located, likely at the edge
							isActive = false
						}
					}
				}
				numSolvedCells++
				n = 0
				if pits[rowN][colN] {
					n = flatindex + 1
				}
				gc = newGridCell(rowN, colN, n)
				s = streams.Value(rowN-1, colN-1)
				if s != streamsNodata && s > 0 {
					isStream = true
				} else {
					isStream = false
				}
				if isStream {
					p = int64(int64(zN*elevMultiplier)*10000 + (int64(n) % 10000))
				} else {
					p = int64(10000000000000 + int64(zN*elevMultiplier)*10000 + (int64(n) % 10000))
				}
				pq.Push(gc, p)
				inQueue[rowN][colN] = true
			}
		}
		progress = int(100.0 * numSolvedCells / numCellsTotal)
		if progress != oldProgress {
			printf("\rBreaching DEM (2 of 2): %v%%", progress)
			oldProgress = progress
		}
	}

	// output the data
	config := raster.NewDefaultRasterConfig()
	config.PreferredPalette = paletteName
	config.DataType = raster.DT_FLOAT32
	config.NoDataValue = nodata
	displayMin := demConfig.DisplayMinimum
	displayMax := demConfig.DisplayMaximum
	config.CoordinateRefSystemWKT = demConfig.CoordinateRefSystemWKT
	config.EPSGCode = demConfig.EPSGCode
	rout, err := raster.CreateNewRaster(this.outputFile, rows, columns,
		dem.North, dem.South, dem.East, dem.West, config)
	if err != nil {
		panic("Failed to write raster")
	}

	printf("\nSaving DEM data...\n")
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			//			s = streams.Value(row, col)
			//			if s != streamsNodata && s > 0 && output[row+1][col+1] != nodata {
			//				z = output[row+1][col+1] - SMALL_NUM*2
			//			} else {
			//				z = output[row+1][col+1]
			//			}
			z = output[row+1][col+1]
			rout.SetValue(row, col, z)
		}
	}

	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by BreachStreams tool"))
	config.DisplayMinimum = displayMin
	config.DisplayMaximum = displayMax
	rout.SetRasterConfig(config)
	rout.Save()

	println("Operation complete!")

	value := fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	println(value)

	overallTime := time.Since(start1)
	value = fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)

	if numUnsolvedPits > 0 {
		printf("Num. of unbreached pits/flats: %v (%f%% of total)\n", numUnsolvedPits, (100.0 * float64(numUnsolvedPits) / float64(numSolvedCells)))
	} else {
		println("All pits/flats were resolved by breaching")
	}
}
