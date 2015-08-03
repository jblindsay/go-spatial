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
	"gospatial/structures"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type BreachDepressions struct {
	inputFile            string
	outputFile           string
	maxLength            int32
	maxDepth             float64
	constrainedBreaching bool
	postBreachFilling    bool
	toolManager          *PluginToolManager
}

func (this *BreachDepressions) GetName() string {
	s := "BreachDepressions"
	return getFormattedToolName(s)
}

func (this *BreachDepressions) GetDescription() string {
	s := "Removes depressions in DEMs using selective breaching"
	return getFormattedToolDescription(s)
}

func (this *BreachDepressions) GetHelpDocumentation() string {
	ret := "This tool is used to remove the sinks (i.e. topographic depressions and flat areas) from digital elevation models (DEMs) using a highly efficient and flexible breaching, or carving, method."
	return ret
}

func (this *BreachDepressions) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

// Can be called to gather a listing of the arguments required to run this tool.
func (this *BreachDepressions) GetArgDescriptions() [][]string {
	numArgs := 6
	ret := structures.Create2dStringArray(numArgs, 3)

	ret[0][0] = "InputDEM"
	ret[0][1] = "string"
	ret[0][2] = "The input DEM name with file extension"

	ret[1][0] = "OutputFile"
	ret[1][1] = "string"
	ret[1][2] = "The output filename with file extension"

	ret[2][0] = "MaxDepth"
	ret[2][1] = "float64"
	ret[2][2] = "The maximum breach channel depth (-1 to ignore)"

	ret[3][0] = "MaxLength"
	ret[3][1] = "int"
	ret[3][2] = "The maximum length of a breach channel (-1 to ignore)"

	ret[4][0] = "ConstrainedBreaching"
	ret[4][1] = "bool"
	ret[4][2] = "Use constrained breaching?"

	ret[5][0] = "SubsequentFilling"
	ret[5][1] = "bool"
	ret[5][2] = "Perform post-breach filling?"

	return ret
}

// ParseArguments is used when the tool is run using command-line args
// rather than in interactive input/output mode.
func (this *BreachDepressions) ParseArguments(args []string) {
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

	if len(strings.TrimSpace(args[2])) > 0 && args[2] != "not specified" {
		if maxDepth, err := strconv.ParseFloat(strings.TrimSpace(args[2]), 64); err == nil {
			this.maxDepth = maxDepth
			if this.maxDepth < 0 {
				this.maxDepth = math.MaxFloat32
			}
		} else {
			this.maxDepth = math.MaxFloat32
			println(err)
		}
	} else {
		this.maxDepth = math.MaxFloat32
	}
	if len(strings.TrimSpace(args[3])) > 0 && args[3] != "not specified" {
		if maxLength, err := strconv.ParseFloat(strings.TrimSpace(args[3]), 64); err == nil {
			this.maxLength = int32(maxLength)
			if this.maxLength < 0 {
				this.maxLength = math.MaxInt32
			}
		} else {
			this.maxLength = math.MaxInt32
			println(err)
		}
	} else {
		this.maxLength = math.MaxInt32
	}

	this.constrainedBreaching = false
	if len(strings.TrimSpace(args[4])) > 0 && args[4] != "not specified" {
		var err error
		if this.constrainedBreaching, err = strconv.ParseBool(strings.TrimSpace(args[4])); err != nil {
			this.constrainedBreaching = false
			println(err)
		}
	} else {
		this.constrainedBreaching = false
	}

	this.Run()
}

func (this *BreachDepressions) CollectArguments() {
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

	// get the maxDepth argument
	print("Enter the maximum breach depth (z units): ")
	maxDepthStr, err := consolereader.ReadString('\n')
	if err != nil {
		this.maxDepth = math.MaxFloat64
		println(err)
	}

	if len(strings.TrimSpace(maxDepthStr)) > 0 {
		if this.maxDepth, err = strconv.ParseFloat(strings.TrimSpace(maxDepthStr), 64); err != nil {
			this.maxDepth = math.MaxFloat64
			println(err)
		}
	} else {
		this.maxDepth = -1
	}

	// get the maxDepth argument
	print("Enter the maximum breach channel length (grid cells): ")
	maxLengthStr, err := consolereader.ReadString('\n')
	if len(strings.TrimSpace(maxLengthStr)) > 0 {
		if err != nil {
			this.maxLength = math.MaxInt32
			println(err)
		}
		if maxLength, err := strconv.ParseFloat(strings.TrimSpace(maxLengthStr), 64); err == nil {
			this.maxLength = int32(maxLength)
		} else {
			this.maxLength = math.MaxInt32
			println(err)
		}
	} else {
		this.maxLength = -1
	}

	// get the constrained breaching argument
	print("Use constrained breaching (T or F)? ")
	constrainedStr, err := consolereader.ReadString('\n')
	if err != nil {
		this.constrainedBreaching = false
		println(err)
	}

	if len(strings.TrimSpace(constrainedStr)) > 0 {
		if this.constrainedBreaching, err = strconv.ParseBool(strings.TrimSpace(constrainedStr)); err != nil {
			this.constrainedBreaching = false
			println(err)
		}
	} else {
		this.constrainedBreaching = false
	}

	if this.maxDepth < math.MaxFloat64 && this.maxLength < math.MaxInt32 {
		print("Perform post-breach filling (T or F)? ")
		postBreachFillStr, err := consolereader.ReadString('\n')
		if err != nil {
			this.postBreachFilling = false
			println(err)
		}

		if len(strings.TrimSpace(postBreachFillStr)) > 0 {
			if this.postBreachFilling, err = strconv.ParseBool(strings.TrimSpace(postBreachFillStr)); err != nil {
				this.postBreachFilling = false
				println(err)
			}
		} else {
			this.postBreachFilling = false
		}
	}

	this.Run()
}

func (this *BreachDepressions) Run() {
	//	cfg := profile.Config{
	//		CPUProfile:     true,
	//		NoShutdownHook: true, // do not hook SIGINT
	//		ProfilePath:    "/Users/johnlindsay/Documents/",
	//	}
	//	//profile.Config.ProfilePath("/Users/johnlindsay/Documents/")
	//	prf := profile.Start(&cfg)
	//	defer prf.Stop()

	//this.postBreachFilling = false

	if this.toolManager.BenchMode {
		benchmarkBreachDepressions(this)
		return
	}

	start1 := time.Now()

	var progress, oldProgress, col, row, i, n int
	var colN, rowN, r, c, flatindex int
	numSolvedCells := 0
	var dir byte
	needsFilling := false
	var z, zN, lowestNeighbour float64
	var zTest, zN2 float64
	var gc gridCell
	var p int64
	var breachDepth, maxPathBreachDepth float64
	var numCellsInPath int32
	var isPit, isEdgeCell bool
	numPits := 0
	numPitsSolved := 0
	numUnsolvedPits := 0
	numValidCells := 0
	var isActive bool
	dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
	dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
	backLink := [8]byte{5, 6, 7, 8, 1, 2, 3, 4}
	//outPointer := [9]float64{0, 1, 2, 4, 8, 16, 32, 64, 128}
	maxLengthOrDepthUsed := false
	if this.maxDepth > 0 || this.maxLength > 0 {
		maxLengthOrDepthUsed = true
	}
	if maxLengthOrDepthUsed && this.maxDepth == -1 {
		this.maxDepth = math.MaxFloat64
	}
	if maxLengthOrDepthUsed && this.maxLength == -1 {
		this.maxLength = math.MaxInt32
	}
	performConstrainedBreaching := this.constrainedBreaching
	if !maxLengthOrDepthUsed && performConstrainedBreaching {
		performConstrainedBreaching = false
	}
	//outputPointer := false
	//performFlowAccumulation := false
	println("Reading DEM data...")
	dem, err := raster.CreateRasterFromFile(this.inputFile)
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

	//	output := structures.Create2dFloat64Array(rows+2, columns+2)
	//	pits := structures.Create2dBoolArray(rows+2, columns+2)
	//	inQueue := structures.Create2dBoolArray(rows+2, columns+2)
	//	flowdir := structures.Create2dByteArray(rows+2, columns+2)

	pq := NewPQueue()

	//q := NewQueue()
	var floodorder []int
	//floodorder := make([]int, numCellsTotal)
	floodOrderTail := 0

	// find the pit cells and initialize the grids
	printf("\rBreaching DEM (1 of 2): %v%%", 0)
	oldProgress = 0
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			z = dem.Value(row, col)
			output[row+1][col+1] = z
			flowdir[row+1][col+1] = 0
			if z != nodata {
				isPit = true
				isEdgeCell = false
				lowestNeighbour = POS_INF
				for n = 0; n < 8; n++ {
					zN = dem.Value(row+dY[n], col+dX[n])
					if zN != nodata && zN < z {
						isPit = false
						break
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
					p = int64(int64(z*elevMultiplier) * 100000)
					pq.Push(gc, p)
					inQueue[row+1][col+1] = true
				}
				if isPit {
					//					if isEdgeCell { // pit on an edge
					//						gc = newGridCell(row+1, col+1, 0)
					//						p = int64(int64(z*elevMultiplier) * 100000)
					//						//						item = &Item{
					//						//							value:    gc,
					//						//							priority: p,
					//						//						}
					//						//						heap.Push(&pq, item)
					//						pq.Push(gc, p)
					//						inQueue[row+1][col+1] = true
					//					} else { // interior pit
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
					if lowestNeighbour != POS_INF {
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

	//heap.Init(&pq)

	// now breach
	printf("\r                                                                 ")
	oldProgress = int(100.0 * numSolvedCells / numCellsTotal)
	printf("\rBreaching DEM (2 of 2): %v%%", oldProgress)

	if !maxLengthOrDepthUsed {
		// Perform a complete breaching solution; there will be no subseqent filling
		for numPitsSolved < numPits {
			gc = pq.Pop()
			row = gc.row
			col = gc.column
			flatindex = gc.flatIndex
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
					p = int64(int64(zN*elevMultiplier)*100000 + (int64(n) % 100000))
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
	} else if !performConstrainedBreaching {
		// Perform selective breaching. Sinks that can be removed within the
		// specified constraints of the max breach length and depth will
		// be breached. Otherwise they will be removed during a subsequent
		// filling operation.
		floodorder = make([]int, numValidCells)
		for pq.Len() > 0 { //numPitsSolved < numPits {
			gc = pq.Pop()
			row = gc.row
			col = gc.column
			if this.postBreachFilling {
				//q.Push(row, col)
				floodorder[floodOrderTail] = row*columns + col
				floodOrderTail++
			}
			flatindex = gc.flatIndex
			for i = 0; i < 8; i++ {
				rowN = row + dY[i]
				colN = col + dX[i]
				zN = output[rowN][colN]
				if zN != nodata && !inQueue[rowN][colN] {
					flowdir[rowN][colN] = backLink[i]
					if pits[rowN][colN] {
						numPitsSolved++
						// trace the flowpath back until you find a lower cell
						// or a constraint is encountered
						numCellsInPath = 0
						maxPathBreachDepth = 0

						zTest = zN
						r = rowN
						c = colN
						isActive = true
						for isActive {
							zTest -= SMALL_NUM // ensures a small increment slope
							dir = flowdir[r][c]
							if dir > 0 {
								r += dY[dir-1]
								c += dX[dir-1]
								zN2 = output[r][c]
								if zN2 <= zTest || zN2 == nodata {
									// a lower grid cell has been found
									isActive = false
								} else {
									breachDepth = dem.Value(r-1, c-1) - zTest
									if breachDepth > maxPathBreachDepth {
										maxPathBreachDepth = breachDepth
									}
								}
							} else {
								isActive = false
							}
							numCellsInPath++
							if numCellsInPath > this.maxLength {
								isActive = false
							}
							if maxPathBreachDepth > this.maxDepth {
								isActive = false
							}
						}

						if numCellsInPath <= this.maxLength && maxPathBreachDepth <= this.maxDepth {
							// breach it completely
							zTest = zN
							r = rowN
							c = colN
							isActive = true
							for isActive {
								zTest -= SMALL_NUM // ensures a small increment slope
								dir = flowdir[r][c]
								if dir > 0 {
									r += dY[dir-1]
									c += dX[dir-1]
									zN2 = output[r][c]
									if zN2 <= zTest || zN2 == nodata {
										// a lower grid cell has been found
										isActive = false
									} else {
										output[r][c] = zTest
									}
								} else {
									isActive = false
								}
							}
						} else {
							// it will be removed by filling in the next step.
							needsFilling = true
							numUnsolvedPits++
						}
					}
					numSolvedCells++
					n = 0
					if pits[rowN][colN] {
						n = flatindex + 1
					}
					gc = newGridCell(rowN, colN, n)
					p = int64(int64(zN*elevMultiplier)*100000 + (int64(n) % 100000))
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
	} else {
		// perform constrained breaching
		floodorder = make([]int, numValidCells)
		var outletHeight float64
		var outletDist, targetDist, j int32
		var zOrig float64
		for pq.Len() > 0 { //numPitsSolved < numPits {
			//item := heap.Pop(&pq).(*Item)
			//gc = item.value
			gc = pq.Pop()
			row = gc.row
			col = gc.column
			if this.postBreachFilling {
				//q.Push(row, col)
				floodorder[floodOrderTail] = row*columns + col
				floodOrderTail++
			}
			flatindex = gc.flatIndex
			//z = output[row][col]
			for i = 0; i < 8; i++ {
				rowN = row + dY[i]
				colN = col + dX[i]
				zN = output[rowN][colN]
				if zN != nodata && !inQueue[rowN][colN] {
					flowdir[rowN][colN] = backLink[i]
					if pits[rowN][colN] {
						numPitsSolved++
						// trace the flowpath back until you find a lower cell
						// or a constraint is encountered
						numCellsInPath = 0
						maxPathBreachDepth = 0

						zTest = zN
						r = rowN
						c = colN
						outletHeight = -math.MaxFloat64
						outletDist = 0
						isActive = true
						for isActive {
							zTest -= SMALL_NUM // ensures a small increment slope
							dir = flowdir[r][c]
							if dir > 0 {
								r += dY[dir-1]
								c += dX[dir-1]
								zN2 = output[r][c]
								if zN2 <= zTest || zN2 == nodata {
									// a lower grid cell has been found
									isActive = false
								} else {
									zOrig = dem.Value(r-1, c-1)
									breachDepth = zOrig - zTest
									if breachDepth > maxPathBreachDepth {
										maxPathBreachDepth = breachDepth
									}
									if zOrig > outletHeight {
										outletHeight = zOrig
										outletDist = numCellsInPath
									}
								}
							} else {
								isActive = false
							}
							numCellsInPath++
						}

						if numCellsInPath <= this.maxLength && maxPathBreachDepth <= this.maxDepth {
							// breach it completely
							zTest = zN
							r = rowN
							c = colN
							isActive = true
							for isActive {
								zTest -= SMALL_NUM // ensures a small increment slope
								dir = flowdir[r][c]
								if dir > 0 {
									r += dY[dir-1]
									c += dX[dir-1]
									zN2 = output[r][c]
									if zN2 <= zTest || zN2 == nodata {
										// a lower grid cell has been found
										isActive = false
									} else {
										output[r][c] = zTest
									}
								} else {
									isActive = false
								}
							}
						} else {
							// ***Constrained Breaching***
							// it will be completely removed by filling in the next step...
							needsFilling = true
							// but in the meantime, lower the outlet as much as you can.

							zTest = outletHeight - this.maxDepth
							targetDist = numCellsInPath

							if numCellsInPath > this.maxLength {
								if outletDist < this.maxLength/2 {
									targetDist = this.maxLength
								} else {
									targetDist = outletDist + this.maxLength/2
								}
								r = rowN
								c = colN
								for j = 0; j < targetDist; j++ {
									dir = flowdir[r][c]
									if dir > 0 {
										r += dY[dir-1]
										c += dX[dir-1]
										zTest = output[r][c]
									} else {
										break
									}
								}
								if outletHeight-zTest > this.maxDepth {
									zTest = outletHeight - this.maxDepth
								}
							}

							r = rowN
							c = colN
							isActive = true
							numCellsInPath = 0
							for isActive {
								dir = flowdir[r][c]
								if dir > 0 {
									r += dY[dir-1]
									c += dX[dir-1]
									zN2 = output[r][c]
									if zN2 <= zN || zN2 == nodata {
										// a lower grid cell has been found
										isActive = false
									} else {
										if output[r][c] > zTest {
											output[r][c] = zTest
										}
									}
								} else {
									isActive = false
								}
								numCellsInPath++
								if numCellsInPath > targetDist {
									isActive = false
								}
							}
						}
					}
					numSolvedCells++
					n = 0
					if pits[rowN][colN] {
						n = flatindex + 1
					}
					gc = newGridCell(rowN, colN, n)
					p = int64(int64(zN*elevMultiplier)*100000 + (int64(n) % 100000))
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
	}

	pits = nil
	inQueue = nil

	if needsFilling && this.postBreachFilling {
		// Fill the DEM.
		printf("\r                                                                ")

		numSolvedCells = 0
		//for q.Len() > 0 {
		//row, col = q.Pop()
		for c := 0; c < numValidCells; c++ {
			row = floodorder[c] / columns
			col = floodorder[c] % columns
			if row >= 0 && col >= 0 {
				z = output[row][col]
				dir = flowdir[row][col]
				if dir > 0 {
					rowN = row + dY[dir-1]
					colN = col + dX[dir-1]
					zN = output[rowN][colN]
					if zN != nodata {
						if z <= zN+SMALL_NUM {
							output[row][col] = zN + SMALL_NUM
						}
					}
				}
			}
			numSolvedCells++
			progress = int(100.0 * numSolvedCells / numValidCells)
			if progress != oldProgress {
				printf("\rFilling DEM: %v%%", progress)
				oldProgress = progress
			}
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
			z = output[row+1][col+1]
			rout.SetValue(row, col, z)
		}
	}

	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start2)
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by BreachDepressions tool"))
	rout.AddMetadataEntry(fmt.Sprintf("Max breach depth: %v", this.maxDepth))
	rout.AddMetadataEntry(fmt.Sprintf("Max breach length: %v", this.maxLength))
	rout.AddMetadataEntry(fmt.Sprintf("Constrained Breaching: %v", this.constrainedBreaching))
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

type gridCell struct {
	row       int
	column    int
	flatIndex int
}

func newGridCell(r, c, f int) (gc gridCell) {
	gc = gridCell{r, c, f}
	return gc
}

// An Item is something we manage in a priority queue.
//type Item struct {
//	value    gridCell // The value of the item; arbitrary.
//	priority int64    // The priority of the item in the queue.
//	// The index is needed by update and is maintained by the heap.Interface methods.
//	index int // The index of the item in the heap.
//}

// A PriorityQueue implements heap.Interface and holds Items.
//type PriorityQueue []*Item

//func (pq PriorityQueue) Len() int { return len(pq) }

//func (pq PriorityQueue) Less(i, j int) bool {
//	return pq[i].priority < pq[j].priority
//}

//func (pq PriorityQueue) Swap(i, j int) {
//	pq[i], pq[j] = pq[j], pq[i]
//	pq[i].index = i
//	pq[j].index = j
//}

//func (pq *PriorityQueue) Push(x interface{}) {
//	n := len(*pq)
//	item := x.(*Item)
//	item.index = n
//	*pq = append(*pq, item)
//}

//func (pq *PriorityQueue) Pop() interface{} {
//	old := *pq
//	n := len(old)
//	item := old[n-1]
//	item.index = -1 // for safety
//	*pq = old[0 : n-1]
//	return item
//}

// update modifies the priority and value of an Item in the queue.
//func (pq *PriorityQueue) update(item *Item, value gridCell, priority int64) {
//	item.value = value
//	item.priority = priority
//	heap.Fix(pq, item.index)
//}

type item struct {
	value    gridCell
	priority int64
}

// PQueue is a heap priority queue data structure implementation.
type PQueue struct {
	items      []*item
	elemsCount int
}

func newItem(value gridCell, priority int64) *item {
	return &item{
		value:    value,
		priority: priority,
	}
}

// NewPQueue creates a new priority queue
func NewPQueue() *PQueue {
	items := make([]*item, 1)
	items[0] = nil // Heap queue first element should always be nil

	return &PQueue{
		items:      items,
		elemsCount: 0,
	}
}

func appendItem(slice []*item, data *item) []*item {
	m := len(slice)
	n := m + 1
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]*item, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	slice[m] = data
	//copy(slice[m:n], data)
	return slice
}

// Push the value item into the priority queue with provided priority.
func (pq *PQueue) Push(value gridCell, priority int64) {
	item := newItem(value, priority)

	//pq.items = append(pq.items, item)
	pq.items = appendItem(pq.items, item)
	pq.elemsCount += 1
	pq.swim(pq.elemsCount)
}

// Pop and returns the highest priority item
func (pq *PQueue) Pop() gridCell {
	var max *item = pq.items[1]

	pq.items[1], pq.items[pq.elemsCount] = pq.items[pq.elemsCount], pq.items[1]
	pq.items = pq.items[0:pq.elemsCount]
	pq.elemsCount -= 1
	pq.sink(1)

	return max.value
}

func (pq *PQueue) Len() int {
	return pq.elemsCount
}

func (pq *PQueue) swim(k int) {
	for k > 1 && (pq.items[k/2].priority > pq.items[k].priority) {
		pq.items[k/2], pq.items[k] = pq.items[k], pq.items[k/2]
		k = k / 2
	}
}

func (pq *PQueue) sink(k int) {
	var j int
	for 2*k <= pq.elemsCount {
		j = 2 * k

		if j < pq.elemsCount && (pq.items[j].priority > pq.items[j+1].priority) {
			j++
		}

		if !(pq.items[k].priority > pq.items[j].priority) {
			break
		}

		pq.items[k], pq.items[j] = pq.items[j], pq.items[k]
		k = j
	}
}

// Queue data struture
type queuenode struct {
	row    int
	column int
	next   *queuenode
}

////	A FIFO (first in first out) data stucture.
//type Queue struct {
//	head  *queuenode
//	tail  *queuenode
//	count int
//}

////	Creates a new pointer to a new queue.
//func NewQueue() *Queue {
//	q := &Queue{}
//	return q
//}

////	Returns the number of elements in the queue (i.e. size/length)
//func (q *Queue) Len() int {
//	return q.count
//}

////	Pushes/inserts a value at the end/tail of the queue.
//func (q *Queue) Push(row, column int) {
//	n := &queuenode{row: row, column: column}

//	if q.tail == nil {
//		q.tail = n
//		q.head = n
//	} else {
//		q.tail.next = n
//		q.tail = n
//	}
//	q.count++
//}

////	Returns the value at the front of the queue.
////	i.e. the oldest value in the queue.
//func (q *Queue) Pop() (int, int) {
//	if q.head == nil {
//		return -1, -1
//	}

//	n := q.head
//	q.head = n.next

//	if q.head == nil {
//		q.tail = nil
//	}
//	q.count--

//	return n.row, n.column
//}
