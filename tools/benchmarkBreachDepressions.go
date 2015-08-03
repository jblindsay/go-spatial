// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// March. 2015.

package tools

import (
	"math"
	"strconv"
	"time"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

/* This function is only used to benchmark the BreachDepressions tool.
      It can be called by running the tool in 'benchon' mode. The tool is run
	10 times and elapsed times do not include disk I/O. No output file
	is created.
*/
func benchmarkBreachDepressions(parent *BreachDepressions) {
	println("Benchmarking BreachDepressions...")

	var progress, oldProgress, col, row, i, n int
	var colN, rowN, r, c, flatindex int
	var dir byte
	needsFilling := false
	var z, zN, lowestNeighbour float64
	var zTest, zN2 float64
	var gc gridCell
	var p int64
	var breachDepth, maxPathBreachDepth float64
	var numCellsInPath int32
	var isPit, isEdgeCell bool
	var isActive bool
	dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
	dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
	backLink := [8]byte{5, 6, 7, 8, 1, 2, 3, 4}
	//outPointer := [9]float64{0, 1, 2, 4, 8, 16, 32, 64, 128}
	maxLengthOrDepthUsed := false
	if parent.maxDepth > 0 || parent.maxLength > 0 {
		maxLengthOrDepthUsed = true
	}
	if maxLengthOrDepthUsed && parent.maxDepth == -1 {
		parent.maxDepth = math.MaxFloat64
	}
	if maxLengthOrDepthUsed && parent.maxLength == -1 {
		parent.maxLength = math.MaxInt32
	}
	performConstrainedBreaching := parent.constrainedBreaching
	if !maxLengthOrDepthUsed && performConstrainedBreaching {
		performConstrainedBreaching = false
	}
	println("Reading DEM data...")
	dem, err := raster.CreateRasterFromFile(parent.inputFile)
	if err != nil {
		println(err.Error())
	}
	rows := dem.Rows
	columns := dem.Columns
	rowsLessOne := rows - 1
	numCellsTotal := rows * columns
	nodata := dem.NoDataValue
	minVal := dem.GetMinimumValue()
	elevDigits := len(strconv.Itoa(int(dem.GetMaximumValue() - minVal)))
	elevMultiplier := math.Pow(10, float64(8-elevDigits))
	SMALL_NUM := 1 / elevMultiplier
	POS_INF := math.Inf(1)

	println("The tool will now be run 10 times...")
	var benchTimes [10]time.Duration
	for bt := 0; bt < 10; bt++ {

		println("Run", (bt + 1), "...")

		startTime := time.Now()

		numSolvedCells := 0
		numPits := 0
		numPitsSolved := 0
		numValidCells := 0

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

		//		output := structures.Create2dFloat64Array(rows+2, columns+2)
		//		pits := structures.Create2dBoolArray(rows+2, columns+2)
		//		inQueue := structures.Create2dBoolArray(rows+2, columns+2)
		//		flowdir := structures.Create2dByteArray(rows+2, columns+2)

		pq := NewPQueue()
		//floodorder := NewQueue()
		var floodorder []int
		//floodorder := make([]int, numCellsTotal)
		floodOrderTail := 0

		// find the pit cells and initialize the grids
		printf("\rBreaching DEM (1 of 2): %v%%", 0)
		oldProgress = 0
		for row = 0; row < rows; row++ {
			for col = 0; col < columns; col++ {
				z = dem.Value(row, col) // input[row+1][col+1]
				output[row+1][col+1] = z
				flowdir[row+1][col+1] = 0
				if z != nodata {
					isPit = true
					isEdgeCell = false
					lowestNeighbour = POS_INF
					for n = 0; n < 8; n++ {
						zN = dem.Value(row+dY[n], col+dX[n]) //input[row+dY[n]+1][col+dX[n]+1]
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
			for pq.Len() > 0 {
				gc = pq.Pop()
				row = gc.row
				col = gc.column
				if parent.postBreachFilling {
					//floodorder.Push(row, col)
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
										breachDepth = dem.Value(r-1, c-1) - zTest //input[r][c] - zTest
										if breachDepth > maxPathBreachDepth {
											maxPathBreachDepth = breachDepth
										}
									}
								} else {
									isActive = false
								}
								numCellsInPath++
								if numCellsInPath > parent.maxLength {
									isActive = false
								}
								if maxPathBreachDepth > parent.maxDepth {
									isActive = false
								}
							}

							if numCellsInPath <= parent.maxLength && maxPathBreachDepth <= parent.maxDepth {
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
			floodorder = make([]int, numValidCells)
			// perform constrained breaching
			var outletHeight float64
			var outletDist, targetDist, j int32
			var zOrig float64
			for pq.Len() > 0 {
				//item := heap.Pop(&pq).(*Item)
				//gc = item.value
				gc = pq.Pop()
				row = gc.row
				col = gc.column
				if parent.postBreachFilling {
					//floodorder.Push(row, col)
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
										zOrig = dem.Value(r-1, c-1) //input[r][c]
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

							if numCellsInPath <= parent.maxLength && maxPathBreachDepth <= parent.maxDepth {
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

								zTest = outletHeight - parent.maxDepth
								targetDist = numCellsInPath

								if numCellsInPath > parent.maxLength {
									if outletDist < parent.maxLength/2 {
										targetDist = parent.maxLength
									} else {
										targetDist = outletDist + parent.maxLength/2
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
									if outletHeight-zTest > parent.maxDepth {
										zTest = outletHeight - parent.maxDepth
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

		if parent.postBreachFilling && needsFilling {
			// Fill the DEM.
			printf("\r                                                                    ")

			numSolvedCells = 0
			//for numSolvedCells < numCellsTotal {
			//	row, col = floodorder.Pop()
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

		benchTimes[bt] = time.Since(startTime)
		printf("     Elapsed time (s): %v\n", benchTimes[bt].Seconds())
	}
	println("")
	println("Elapsed times (in sec.) of the 10 runs:")
	avgVal := 0.0
	for i := 0; i < 10; i++ {
		println(benchTimes[i].Seconds())
		avgVal += benchTimes[i].Seconds()
	}

	println("Average Time: ", avgVal/10.0)

	println("Operation complete!")

}
