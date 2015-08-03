// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// March. 2015.

package tools

import (
	"fmt"
	"gospatial/geospatialfiles/raster"
	"math"
	"strconv"
	"time"
)

/* This function is only used to benchmark the BreachDepressions tool.
      It can be called by running the tool in 'benchon' mode. The tool is run
	10 times and elapsed times do not include disk I/O. No output file
	is created.
*/
func benchmarkFillDepressions(parent *FillDepressions) {
	println("Benchmarking FillDepressions...")

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
	dem, err := raster.CreateRasterFromFile(parent.inputFile)
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
	rout, err := raster.CreateNewRaster(parent.outputFile, rows, columns,
		dem.North, dem.South, dem.East, dem.West, config)
	if err != nil {
		panic("Failed to write raster")
	}

	minVal := dem.GetMinimumValue()
	elevDigits := len(strconv.Itoa(int(dem.GetMaximumValue() - minVal)))
	elevMultiplier := math.Pow(10, float64(8-elevDigits))
	SMALL_NUM := 1 / elevMultiplier
	if !parent.fixFlats {
		SMALL_NUM = 0
	}

	println("The tool will now be run 10 times...")
	var benchTimes [10]time.Duration
	for bt := 0; bt < 10; bt++ {

		println("Run", (bt + 1), "...")

		startTime := time.Now()

		// Fill the DEM.
		inQueue := make([][]bool, rows+2)

		for i = 0; i < rows+2; i++ {
			inQueue[i] = make([]bool, columns+2)
		}

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
						}
					}

					if isEdgeCell {
						gc = newGridCell(row, col, 0)
						p = int64(int64(zN*elevMultiplier) * 100000)
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

		printf("\r                                                      ")
		oldProgress = -1
		for pq.Len() > 0 {
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
