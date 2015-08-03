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
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FD8FlowAccum struct {
	inputFile   string
	outputFile  string
	lnTransform bool
	power       float32
	parallel    bool
	toolManager *PluginToolManager
}

func (this *FD8FlowAccum) GetName() string {
	s := "FD8FlowAccum"
	return getFormattedToolName(s)
}

func (this *FD8FlowAccum) GetDescription() string {
	s := "Performs FD8 flow accumulation on a DEM"
	return getFormattedToolDescription(s)
}

func (this *FD8FlowAccum) GetHelpDocumentation() string {
	ret := "This tool calculates a FD8 flow accumulation raster from a digital elevation model (DEM)."
	return ret
}

func (this *FD8FlowAccum) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *FD8FlowAccum) GetArgDescriptions() [][]string {
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

	ret[2][0] = "LogTransform"
	ret[2][1] = "bool"
	ret[2][2] = "Log transform the output?"

	ret[3][0] = "PerformParallel"
	ret[3][1] = "bool"
	ret[3][2] = "Perform the analysis in parallel?"

	return ret
}

func (this *FD8FlowAccum) ParseArguments(args []string) {
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

	this.parallel = false
	if len(strings.TrimSpace(args[3])) > 0 && args[3] != "not specified" {
		var err error
		if this.parallel, err = strconv.ParseBool(strings.TrimSpace(args[3])); err != nil {
			this.parallel = false
			println(err)
		}
	} else {
		this.parallel = false
	}
	this.Run()
}

func (this *FD8FlowAccum) CollectArguments() {
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

	// get the perform parallel argument
	print("Perform in parallel (T or F)? ")
	parallelStr, err := consolereader.ReadString('\n')
	if err != nil {
		this.parallel = false
		println(err)
	}

	if len(strings.TrimSpace(parallelStr)) > 0 {
		if this.parallel, err = strconv.ParseBool(strings.TrimSpace(parallelStr)); err != nil {
			this.parallel = false
			println(err)
		}
	} else {
		this.parallel = false
	}

	this.Run()
}

func (this *FD8FlowAccum) Run() {
	start1 := time.Now()

	//var z, zN float64
	var progress, oldProgress int
	var col, row int
	//power := 2.0

	println("Reading DEM data...")
	dem, err := raster.CreateRasterFromFile(this.inputFile)
	if err != nil {
		println(err.Error())
	}
	rows := dem.Rows
	columns := dem.Columns
	nodata := dem.NoDataValue
	println("Calculating pointer grid...")

	numCPUs := runtime.NumCPU()

	if numCPUs > 1 && this.parallel {
		numInflowing := structures.NewParallelRectangularArrayByte(rows, columns)
		//numInflowing := structures.NewRectangularArrayByte(rows, columns)

		outputData := structures.NewParallelRectangularArrayFloat64(rows, columns, nodata)
		//outputData := structures.NewRectangularArrayFloat64(rows, columns, nodata)
		//outputData.InitializeWithConstant(1.0)

		// parallel stuff
		println("Num CPUs:", numCPUs)
		c1 := make(chan bool)
		//c2 := make(chan bool)
		runtime.GOMAXPROCS(numCPUs)
		var wg sync.WaitGroup

		qg := NewQueueGroup(numCPUs)

		//		go func(rows, columns) {
		//			numCells := rows * columns
		//			progress, oldProgress := 0, -1
		//			numCellsCompleted := 0
		//			for numCellsCompleted < numCells {
		//				<-c2
		//				numCellsCompleted += increment
		//				if report {
		//					progress = int(100.0 * float64(numCellsCompleted) / float64(numCells))
		//					if progress != oldProgress {
		//						printf("\rLoop (2 of 2): %v%%", progress)
		//						oldProgress = progress
		//					}
		//				}
		//			}
		//		}(rows, columns)

		// calculate flow directions
		printf("\r                                                    ")
		printf("\rLoop (1 of 2): %v%%", 0)
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
				var z, zN float64
				var j byte
				dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
				dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
				for row := rowSt; row <= rowEnd; row++ {
					byteData := make([]byte, columns)
					floatData := make([]float64, columns)
					for col := 0; col < columns; col++ {
						z = dem.Value(row, col)
						if z != nodata {
							j = 0
							for n := 0; n < 8; n++ {
								zN = dem.Value(row+dY[n], col+dX[n])
								if zN > z && zN != nodata {
									j++
								}
							}
							byteData[col] = j
							//numInflowing.SetValue(row, col, j)
							if j == 0 {
								qg.push(row, col, k)
							}
							floatData[col] = 1.0
						} else {
							//c2 <- true // update the number of solved cells
							//outputData.SetValue(row, col, nodata)
							floatData[col] = nodata
						}
					}
					numInflowing.SetRowData(row, byteData)
					outputData.SetRowData(row, floatData)
					c1 <- true // row completed
				}

			}(startingRow, endingRow, k)
			startingRow = endingRow + 1
			k++
		}

		oldProgress = -1
		rowsLessOne := rows - 1
		for rowsCompleted := 0; rowsCompleted < rows; rowsCompleted++ {
			<-c1 // a row has successfully completed
			progress = int(100.0 * float64(rowsCompleted) / float64(rowsLessOne))
			if progress != oldProgress {
				printf("\rLoop (1 of 2): %v%%", progress)
				oldProgress = progress
			}
		}

		wg.Wait()

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
		//var numSolvedCells int32 = 0
		println("")
		println("Performing the flow accumulation...")
		for k := 0; k < numCPUs; k++ {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
				dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
				//var numCellsTotal float64 = float64(rows * columns)
				var faValue float64
				var totalWeights float64
				//var progress, oldProgress int = 0, -1
				var z, zN float64
				var col, row, r, c, n int
				power := 2.0
				for qg.length(k) > 0 {
					row, col = qg.pop(k)
					z = dem.Value(row, col)
					faValue = outputData.Value(row, col)
					// calculate the weights
					totalWeights = 0
					weights := [8]float64{0, 0, 0, 0, 0, 0, 0, 0}
					downslope := [8]bool{false, false, false, false, false, false, false, false}
					for n = 0; n < 8; n++ {
						zN = dem.Value(row+dY[n], col+dX[n])
						if zN < z && zN != nodata {
							weights[n] = math.Pow(z-zN, power)
							totalWeights += weights[n]
							downslope[n] = true
						}
					}

					// now perform the neighbour accumulation
					for n = 0; n < 8; n++ {
						r = row + dY[n]
						c = col + dX[n]
						//zN = dem.Value(r, c)
						if downslope[n] {
							outputData.Increment(r, c, faValue*(weights[n]/totalWeights))
							p := numInflowing.DecrementAndReturn(r, c, 1.0)

							//see if you can progress further downslope
							if p == 0 {
								qg.push(r, c, k)
							}
						}
					}
					//c2 <- true
					//					atomic.AddInt32(&numSolvedCells, 1)
					//					progress = int(100.0 * float64(numSolvedCells) / numCellsTotal)
					//					if progress != oldProgress {
					//						printf("\rLoop (2 of 2): %v%%", progress)
					//						oldProgress = progress
					//					}
				}
			}(k)
		}

		//		oldProgress = -1
		//		for rowsCompleted := 0; rowsCompleted < rows; rowsCompleted++ {
		//			<-c1 // a row has successfully completed
		//			progress = int(100.0 * float64(rowsCompleted) / float64(rowsLessOne))
		//			if progress != oldProgress {
		//				printf("\rLoop (1 of 2): %v%%", progress)
		//				oldProgress = progress
		//			}
		//		}

		wg.Wait()

		if this.lnTransform {
			println("")
			printf("\r                                                    ")
			printf("\rTransforming output: %v%%", 0)
			oldProgress = 0
			//var z float64
			var rowsLessOne int32 = int32(rows - 1)
			for row = 0; row < rows; row++ {
				floatData := outputData.GetRowData(row)
				for col = 0; col < columns; col++ {
					//z = rout.Value(row, col)
					//z = outputData.Value(row, col)
					if floatData[col] != nodata {
						//rout.SetValue(row, col, math.Log(z))
						rout.SetValue(row, col, math.Log(floatData[col]))
					}
				}

				progress = int(100.0 * int32(row) / rowsLessOne)
				if progress != oldProgress {
					printf("\rTransforming output: %v%%", progress)
					oldProgress = progress
				}
			}
		} else {
			println("")
			printf("\r                                                    ")
			printf("\rOutputing data: %v%%", 0)
			oldProgress = 0
			//var z float64
			var rowsLessOne int32 = int32(rows - 1)
			for row = 0; row < rows; row++ {
				floatData := outputData.GetRowData(row)
				for col = 0; col < columns; col++ {
					//z = rout.Value(row, col)
					//z = outputData.Value(row, col)
					//					if floatData[col] != nodata {
					//						rout.SetValue(row, col, z)
					//					} else {
					//						rout.SetValue(row, col, nodata)
					//					}
					rout.SetValue(row, col, floatData[col])
				}
				progress = int(100.0 * int32(row) / rowsLessOne)
				if progress != oldProgress {
					printf("\rOutputing data: %v%%", progress)
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
	} else {
		numInflowing := structures.NewRectangularArrayByte(rows, columns)

		outputData := structures.NewRectangularArrayFloat64(rows, columns, nodata)
		outputData.InitializeWithConstant(1.0)

		q := newQueue()

		// calculate flow directions
		printf("\r                                                    ")
		printf("\rLoop (1 of 2): %v%%", 0)
		var numSolvedCells int32 = 0
		var rowsCompleted int32 = 0
		oldProgress = 0

		var z, zN float64
		var j byte
		var rowsLessOne int32 = int32(rows - 1)
		var progress, oldProgress int32 = 0, -1
		dX := [8]int{1, 1, 1, 0, -1, -1, -1, 0}
		dY := [8]int{-1, 0, 1, 1, 1, 0, -1, -1}
		for row := 0; row <= rows; row++ {
			for col := 0; col < columns; col++ {
				z = dem.Value(row, col)
				if z != nodata {
					j = 0
					for n := 0; n < 8; n++ {
						zN = dem.Value(row+dY[n], col+dX[n])
						if zN > z && zN != nodata {
							j++
						}
					}
					numInflowing.SetValue(row, col, j)
					if j == 0 {
						q.push(row, col)
					}
				} else {
					numSolvedCells++
					outputData.SetValue(row, col, nodata)
				}
			}
			//numInflowing.SetRowData(row, byteData)
			rowsCompleted++
			progress = int32(100.0 * rowsCompleted / rowsLessOne)
			if progress != oldProgress {
				printf("\rLoop (1 of 2): %v%%", progress)
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

		var numCellsTotal float64 = float64(rows * columns)
		var faValue float64
		//var faValueN float64
		var totalWeights float64
		progress, oldProgress = 0, -1
		var col, row, r, c, n int
		power := 2.0
		for q.count > 0 {
			row, col = q.pop()
			z = dem.Value(row, col)
			//faValue = rout.Value(row, col)
			faValue = outputData.Value(row, col)
			// calculate the weights
			totalWeights = 0
			weights := [8]float64{0, 0, 0, 0, 0, 0, 0, 0}
			downslope := [8]bool{false, false, false, false, false, false, false, false}
			for n = 0; n < 8; n++ {
				zN = dem.Value(row+dY[n], col+dX[n])
				if zN < z && zN != nodata {
					weights[n] = math.Pow(z-zN, power)
					totalWeights += weights[n]
					downslope[n] = true
				}
			}

			// now perform the neighbour accumulation
			for n = 0; n < 8; n++ {
				r = row + dY[n]
				c = col + dX[n]
				//zN = dem.Value(r, c)
				if downslope[n] {
					//faValueN = rout.Value(r, c)
					//faValueN = outputData.Value(r, c)
					// update the output grids
					//rout.SetValue(r, c, faValueN+faValue*(weights[n]/totalWeights))
					outputData.Increment(r, c, faValue*(weights[n]/totalWeights))
					numInflowing.Decrement(r, c)

					//see if you can progress further downslope
					//if numInflowing[r+1][c+1] == 0 {
					if numInflowing.Value(r, c) == 0 {
						//qs[k].push(r, c)
						q.push(r, c)
					}
				}
			}

			numSolvedCells++
			progress = int32(100.0 * float64(numSolvedCells) / numCellsTotal)
			if progress != oldProgress {
				printf("\rLoop (2 of 2): %v%%", progress)
				oldProgress = progress
			}
		}

		if this.lnTransform {
			println("")
			printf("\r                                                    ")
			printf("\rTransforming output: %v%%", 0)
			oldProgress = 0
			var z float64
			var rowsLessOne int32 = int32(rows - 1)
			for row = 0; row < rows; row++ {
				for col = 0; col < columns; col++ {
					//z = rout.Value(row, col)
					z = outputData.Value(row, col)
					if z != nodata {
						rout.SetValue(row, col, math.Log(z))
					} else {
						rout.SetValue(row, col, nodata)
					}
				}
				progress = int32(100.0 * int32(row) / rowsLessOne)
				if progress != oldProgress {
					printf("\rTransforming output: %v%%", progress)
					oldProgress = progress
				}
			}
		} else {
			println("")
			printf("\r                                                    ")
			printf("\rOutputing data: %v%%", 0)
			oldProgress = 0
			var z float64
			var rowsLessOne int32 = int32(rows - 1)
			for row = 0; row < rows; row++ {
				for col = 0; col < columns; col++ {
					//z = rout.Value(row, col)
					z = outputData.Value(row, col)
					if z != nodata {
						rout.SetValue(row, col, z)
					} else {
						rout.SetValue(row, col, nodata)
					}
				}
				progress = int32(100.0 * int32(row) / rowsLessOne)
				if progress != oldProgress {
					printf("\rOutputing data: %v%%", progress)
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
	}

	println("Operation complete!")

	overallTime := time.Since(start1)
	value := fmt.Sprintf("Elapsed time (total): %s", overallTime)
	println(value)
}

// Queue data struture
type gridnode struct {
	row    int
	column int
	next   *gridnode
}

//	A thread-safe FIFO (first in first out) data stucture.
type fd8Queue struct {
	head  *gridnode
	tail  *gridnode
	count int
	sync.Mutex
}

//	Creates a new pointer to a new queue.
func newFD8Queue() *fd8Queue {
	q := &fd8Queue{}
	return q
}

//	Returns the number of elements in the queue (i.e. size/length)
func (q *fd8Queue) len() int {
	q.Lock()
	defer q.Unlock()
	return q.count
}

//	Pushes/inserts a value at the end/tail of the queue.
func (q *fd8Queue) push(row, column int) {
	q.Lock()
	defer q.Unlock()
	n := &gridnode{row: row, column: column}

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
func (q *fd8Queue) pop() (int, int) {
	q.Lock()
	defer q.Unlock()
	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}
	q.count--

	return n.row, n.column
}

//	A non-thread-safe FIFO (first in first out) data stucture.
type queue struct {
	head  *gridnode
	tail  *gridnode
	count int
}

//	Creates a new pointer to a new queue.
func newQueue() *queue {
	q := &queue{}
	return q
}

//	Returns the number of elements in the queue (i.e. size/length)
func (q *queue) len() int {
	return q.count
}

//	Pushes/inserts a value at the end/tail of the queue.
func (q *queue) push(row, column int) {
	n := &gridnode{row: row, column: column}

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
func (q *queue) pop() (int, int) {
	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}
	q.count--

	return n.row, n.column
}

type queueGroup struct {
	group     []*queue
	numQueues int
	//lock      bool
}

func NewQueueGroup(numQueues int) *queueGroup {
	qg := &queueGroup{}
	qg.group = make([]*queue, numQueues)
	for i := 0; i < numQueues; i++ {
		qg.group[i] = newQueue()
	}
	qg.numQueues = numQueues
	return qg
}

//	Returns the number of elements in the queue (i.e. size/length)
func (this *queueGroup) length(k int) int {
	//	if this.group[k].count == 0 {
	//		this.lock = true
	//		// see if you can steal work for this thread to do
	//		largestQueue := -1
	//		for i := 0; i < this.numQueues; i++ {
	//			if this.group[i].len() > largestQueue {
	//				largestQueue = i
	//			}
	//		}
	//		largestQueueSize := this.group[largestQueue].len()
	//		if largestQueueSize > 100 {
	//			// steal half the work from this queue
	//			for j := 0; j < largestQueueSize/2; j++ {
	//				row, column := this.group[largestQueue].pop()
	//				this.group[k].push(row, column)
	//			}
	//			//println("\nThread", k, "stole", (largestQueueSize / 2), "entries from thread", largestQueue)
	//		}
	//		this.lock = false
	//	}
	return this.group[k].count
}

//	Pushes/inserts a value at the end/tail of the queue.
func (this *queueGroup) push(row, column, k int) {
	//	for this.lock {
	//		// another thread is currently stealing work so delay any
	//		// modifications to any queue until it's done.
	//	}
	this.group[k].push(row, column)
}

//	Returns the value at the front of the queue.
//	i.e. the oldest value in the queue.
func (this *queueGroup) pop(k int) (int, int) {
	//	for this.lock {
	//		// another thread is currently stealing work so delay any
	//		// modifications to any queue until it's done.
	//	}
	return this.group[k].pop()
}
