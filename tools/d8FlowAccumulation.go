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

type D8FlowAccumulation struct {
	inputFile   string
	outputFile  string
	lnTransform bool
	toolManager *PluginToolManager
}

func (this *D8FlowAccumulation) GetName() string {
	s := "D8FlowAccumulation"
	return getFormattedToolName(s)
}

func (this *D8FlowAccumulation) GetDescription() string {
	s := "Performs D8 flow accumulation on a DEM"
	return getFormattedToolDescription(s)
}

func (this *D8FlowAccumulation) GetHelpDocumentation() string {
	ret := "This tool calculates a D8 flow accumulation raster from a digital elevation model (DEM)."
	return ret
}

func (this *D8FlowAccumulation) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *D8FlowAccumulation) GetArgDescriptions() [][]string {
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

	ret[2][0] = "LogTransform"
	ret[2][1] = "bool"
	ret[2][2] = "Log transform the output?"

	return ret
}

func (this *D8FlowAccumulation) ParseArguments(args []string) {
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

	this.lnTransform = false
	if len(strings.TrimSpace(args[2])) > 0 && args[2] != "not specified" {
		var err error
		if this.lnTransform, err = strconv.ParseBool(strings.TrimSpace(args[2])); err != nil {
			this.lnTransform = false
			println(err)
		}
	} else {
		this.lnTransform = false
	}
	this.Run()
}

func (this *D8FlowAccumulation) CollectArguments() {
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

	// get the ln-transform argument
	print("Log-transform the output (T or F)? ")
	lnTransformStr, err := consolereader.ReadString('\n')
	if err != nil {
		this.lnTransform = false
		println(err)
	}

	if len(strings.TrimSpace(lnTransformStr)) > 0 {
		if this.lnTransform, err = strconv.ParseBool(strings.TrimSpace(lnTransformStr)); err != nil {
			this.lnTransform = false
			println(err)
		}
	} else {
		this.lnTransform = false
	}

	this.Run()
}

func (this *D8FlowAccumulation) Run() {
	start1 := time.Now()

	var z, zN, slope, maxSlope float64
	var progress, oldProgress, col, row, r, c, i, n int
	var dir int8
	//var b int8
	dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
	dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
	//inflowingVals := [8]int8{5, 6, 7, 8, 1, 2, 3, 4}

	println("Reading DEM data...")
	dem, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		println(err.Error())
	}
	rows := dem.Rows
	columns := dem.Columns
	rowsLessOne := rows - 1
	nodata := dem.NoDataValue
	cellSizeX := dem.GetCellSizeX()
	cellSizeY := dem.GetCellSizeY()
	diagDist := math.Sqrt(cellSizeX*cellSizeX + cellSizeY*cellSizeY)
	dist := [8]float64{diagDist, cellSizeX, diagDist, cellSizeY, diagDist, cellSizeX, diagDist, cellSizeY}
	println("Calculating pointer grid...")
	flowdir := make([][]int8, rows+2)
	numInflowing := make([][]int8, rows+2)
	for i = 0; i < rows+2; i++ {
		flowdir[i] = make([]int8, columns+2)
		numInflowing[i] = make([]int8, columns+2)
	}

	// calculate flow directions
	printf("\r                                                    ")
	printf("\rLoop (1 of 3): %v%%", 0)
	oldProgress = 0
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			z = dem.Value(row, col)
			flowdir[row+1][col+1] = 0
			//			numInflowing[row+1][col+1] = 0
			if z != nodata {
				maxSlope = math.Inf(-1)
				for n = 0; n < 8; n++ {
					zN = dem.Value(row+dY[n], col+dX[n])
					if zN != nodata {
						slope = (z - zN) / dist[n]

						if slope > maxSlope {
							maxSlope = slope
							dir = int8(n) + 1
						}
					}
				}
				if maxSlope > 0 {
					flowdir[row+1][col+1] = dir

					// increment the number of inflowing cells for the downslope receiver
					c = col + dX[dir-1] + 1
					r = row + dY[dir-1] + 1
					numInflowing[r][c]++

				} else {
					flowdir[row+1][col+1] = 0
				}
			} else {
				numInflowing[row+1][col+1] = 0
			}
		}
		progress = int(100.0 * row / rowsLessOne)
		if progress != oldProgress {
			printf("\rLoop (1 of 3): %v%%", progress)
			oldProgress = progress
		}
	}

	//	 calculate the number of inflowing neighbours and initialize the flow queue
	//	 with cells with no inflowing neighbours
	fq := newFlowQueue()
	//fq := newQueue(rows * columns / 2)
	numSolvedCells := 0
	println("")
	println("Calculating the number of inflow neighbours...")
	printf("\r                                                    ")
	printf("\rLoop (2 of 3): %v%%", 0)
	oldProgress = 0
	for row = 0; row < rows; row++ {
		for col = 0; col < columns; col++ {
			z = dem.Value(row, col)
			if z != nodata {
				if numInflowing[row+1][col+1] == 0 {
					fq.push(row, col)
				}
			} else {
				numSolvedCells++
			}

		}
		progress = int(100.0 * row / rowsLessOne)
		if progress != oldProgress {
			printf("\rLoop (2 of 3): %v%%", progress)
			oldProgress = progress
		}
	}

	// create the output file
	config := raster.NewDefaultRasterConfig() //dem.GetRasterConfig()
	config.DataType = raster.DT_FLOAT32
	config.NoDataValue = nodata
	config.InitialValue = 1
	config.PreferredPalette = "blueyellow.pal"
	config.CoordinateRefSystemWKT = dem.GetRasterConfig().CoordinateRefSystemWKT
	config.EPSGCode = dem.GetRasterConfig().EPSGCode
	rout, err := raster.CreateNewRaster(this.outputFile, rows, columns,
		dem.North, dem.South, dem.East, dem.West, config)
	if err != nil {
		panic("Failed to write raster")
	}

	// perform the flow accumlation
	println("")
	println("Performing the flow accumulation...")
	numCellsTotal := rows * columns
	oldProgress = -1
	for fq.count > 0 {
		row, col = fq.pop()
		z = rout.Value(row, col)
		//value to send to it's neighbour
		//find it's downslope neighbour
		dir = flowdir[row+1][col+1]
		if dir > 0 {
			col += dX[dir-1]
			row += dY[dir-1]
			r = row + 1
			c = col + 1
			//update the output grids
			zN = rout.Value(row, col)
			rout.SetValue(row, col, zN+z)
			numInflowing[r][c]--
			//see if you can progress further downslope
			if numInflowing[r][c] == 0 {
				//numInflowing[r][c] = -1
				fq.push(row, col)
			}
		}
		numSolvedCells++
		progress = int(100.0 * numSolvedCells / numCellsTotal)
		if progress != oldProgress {
			printf("\rLoop (3 of 3): %v%%", progress)
			oldProgress = progress
		}
	}

	//	// perform the flow accumulation
	//	println("")
	//	println("Performing the flow accumulation...")
	//	printf("\r                                                    ")
	//	printf("\rLoop (3 of 3): %v%%", 0)
	// var trace bool
	//	oldProgress = 0
	//	for row = 0; row < rows; row++ {
	//		for col = 0; col < columns; col++ {
	//			z = dem.Value(row, col)
	//			if z != nodata {
	//				r = row + 1
	//				c = col + 1
	//				if numInflowing[r][c] == 0 {
	//					numInflowing[r][c] = -1
	//					trace = true

	//					for trace {
	//						z = rout.Value(r-1, c-1)
	//						//value to send to it's neighbour
	//						//find it's downslope neighbour
	//						dir = flowdir[r][c]
	//						if dir > 0 {
	//							c += dX[dir-1]
	//							r += dY[dir-1]
	//							//update the output grids
	//							zN = rout.Value(r-1, c-1)
	//							rout.SetValue(r-1, c-1, zN+z)
	//							numInflowing[r][c]--
	//							//see if you can progress further downslope
	//							if numInflowing[r][c] == 0 {
	//								numInflowing[r][c] = -1
	//								trace = true
	//							} else {
	//								trace = false
	//							}
	//						} else {
	//							trace = false
	//						}
	//					}
	//				}
	//			} else {
	//				rout.SetValue(row, col, nodata)
	//			}
	//		}
	//		progress = int(100.0 * row / rowsLessOne)
	//		if progress != oldProgress {
	//			printf("\rLoop (3 of 3): %v%%", progress)
	//			oldProgress = progress
	//		}
	//	}

	if this.lnTransform {
		println("")
		printf("\r                                                    ")
		printf("\rTransforming output: %v%%", 0)
		oldProgress = 0
		for row = 0; row < rows; row++ {
			for col = 0; col < columns; col++ {
				z = rout.Value(row, col)
				if z != nodata {
					rout.SetValue(row, col, math.Log(z))
				}
			}
			progress = int(100.0 * row / rowsLessOne)
			if progress != oldProgress {
				printf("\rTransforming output: %v%%", progress)
				oldProgress = progress
			}
		}
	}

	println("\nSaving data...")
	rout.AddMetadataEntry(fmt.Sprintf("Created on %s", time.Now().Local()))
	elapsed := time.Since(start1)
	rout.AddMetadataEntry(fmt.Sprintf("Elapsed Time: %v", elapsed))
	rout.AddMetadataEntry(fmt.Sprintf("Created by D8FlowAccumulation tool"))
	rout.Save()

	println("Operation complete!")

	//value = fmt.Sprintf("Elapsed time (excluding file I/O): %s", elapsed)
	//println(value)

	overallTime := time.Since(start1)
	value := fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}

// Queue data struture
type flowqueuenode struct {
	row    int
	column int
	next   *flowqueuenode
}

//	A FIFO (first in first out) data stucture.
type flowQueue struct {
	head  *flowqueuenode
	tail  *flowqueuenode
	count int
}

//	Creates a new pointer to a new queue.
func newFlowQueue() *flowQueue {
	q := &flowQueue{}
	return q
}

//	Returns the number of elements in the queue (i.e. size/length)
//func (q *flowQueue) len() int {
//	return q.count
//}

//	Pushes/inserts a value at the end/tail of the queue.
func (q *flowQueue) push(row, column int) {
	n := &flowqueuenode{row: row, column: column}

	if q.count > 0 {
		q.tail.next = n
		q.tail = n
	} else {
		q.tail = n
		q.head = n
	}
	q.count++
}

//	Returns the value at the front of the queue.
//	i.e. the oldest value in the queue.
func (q *flowQueue) pop() (int, int) {
	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}
	q.count--

	return n.row, n.column
}

//type node struct {
//	row    int
//	column int
//}

////type queue []*node
//type queue struct {
//	data []*node
//}

//func newQueue(capacity int) *queue {
//	q := &queue{}
//	q.data = make([]*node, 0, capacity)
//	return q
//}

//func (q *queue) push(row, column int) {
//	n := &node{row: row, column: column}
//	q.data = append(q.data, n)
//}

//func (q *queue) pop() (int, int) {
//	n := (*q).data[0]
//	q.data = q.data[1:]
//	return n.row, n.column
//}

//func (q *queue) len() int {
//	return len(q.data)
//}

//type stack []*node

//func (q *stack) push(n *node) {
//	*q = append(*q, n)
//}

//func (q *stack) pop() (n *node) {
//	x := q.Len() - 1
//	n = (*q)[x]
//	*q = (*q)[:x]
//	return
//}
//func (q *stack) len() int {
//	return len(*q)
//}
