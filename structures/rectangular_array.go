// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// March. 2015.
package structures

import (
	"errors"
	"sync"
)

// This can be used to create a 2d array of float64 type in a way that
// guarantees that the allocations is localized in memory.
func Create2dFloat64Array(rows, columns int) [][]float64 {
	a := make([][]float64, rows)
	e := make([]float64, rows*columns)
	for i := range a {
		a[i] = e[i*columns : (i+1)*columns]
	}
	return a
}

// This can be used to create a 2d array of int type in a way that
// guarantees that the allocations is localized in memory.
func Create2dIntArray(rows, columns int) [][]int {
	a := make([][]int, rows)
	e := make([]int, rows*columns)
	for i := range a {
		a[i] = e[i*columns : (i+1)*columns]
	}
	return a
}

// This can be used to create a 2d array of byte type in a way that
// guarantees that the allocations is localized in memory.
func Create2dByteArray(rows, columns int) [][]byte {
	a := make([][]byte, rows)
	e := make([]byte, rows*columns)
	for i := range a {
		a[i] = e[i*columns : (i+1)*columns]
	}
	return a
}

// This can be used to create a 2d array of bool type in a way that
// guarantees that the allocations is localized in memory.
func Create2dBoolArray(rows, columns int) [][]bool {
	a := make([][]bool, rows)
	e := make([]bool, rows*columns)
	for i := range a {
		a[i] = e[i*columns : (i+1)*columns]
	}
	return a
}

// This can be used to create a 2d array of string type in a way that
// guarantees that the allocations is localized in memory.
func Create2dStringArray(rows, columns int) [][]string {
	a := make([][]string, rows)
	e := make([]string, rows*columns)
	for i := range a {
		a[i] = e[i*columns : (i+1)*columns]
	}
	return a
}

// A rectangular shaped array (matrix) of float64 type. The array is thread-safe.
type RectangularArrayFloat64 struct {
	data          []float64
	rows, columns int
	nodata        float64
}

func NewRectangularArrayFloat64(rows, columns int, nodata float64) *RectangularArrayFloat64 {
	r := RectangularArrayFloat64{rows: rows, columns: columns, nodata: nodata}
	r.data = make([]float64, rows*columns)
	//r.lock = &sync.Mutex{}
	return &r
}

// Returns the number of rows
func (r *RectangularArrayFloat64) GetRows() int {
	return r.rows
}

// Returns the number of columns
func (r *RectangularArrayFloat64) GetColumns() int {
	return r.columns
}

// Returns the nodata value
func (r *RectangularArrayFloat64) GetNodata() float64 {
	return r.nodata
}

// Sets the nodata value
func (r *RectangularArrayFloat64) SetNodata(value float64) {
	r.nodata = value
}

// Retrives an individual cell value in the matrix.
func (r *RectangularArrayFloat64) Value(row, column int) float64 {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		// the row and column are within the bounds of the matrix
		return r.data[row*r.columns+column]
	} else {
		// the row and column are outside the bounds of the matrix
		return r.nodata
	}
}

// Sets an individual cell value in the matrix.
func (r *RectangularArrayFloat64) SetValue(row, column int, value float64) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		r.data[row*r.columns+column] = value
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Returns an entire row of values.
func (r *RectangularArrayFloat64) GetRowData(row int) []float64 {
	values := make([]float64, r.columns)
	for column := 0; column < r.columns; column++ {
		values[column] = r.data[row*r.columns+column]
	}
	return values
}

// Sets and entire row of values.
func (r *RectangularArrayFloat64) SetRowData(row int, values []float64) {
	if row >= 0 && row < r.rows {
		for column := 0; column < r.columns; column++ {
			r.data[row*r.columns+column] = values[column]
		}
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Increments an individual cell value in the matrix.
func (r *RectangularArrayFloat64) Increment(row, column int, values ...float64) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		if len(values) == 0 {
			r.data[row*r.columns+column]++
		} else {
			for _, num := range values {
				r.data[row*r.columns+column] += num
			}
		}
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Decrements an individual cell value in the matrix.
func (r *RectangularArrayFloat64) Decrement(row, column int, values ...float64) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		if len(values) == 0 {
			r.data[row*r.columns+column]--
		} else {
			for _, num := range values {
				r.data[row*r.columns+column] -= num
			}
		}
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Initializes all cells with a constant value.
func (r *RectangularArrayFloat64) InitializeWithConstant(value float64) {
	for i := 0; i < r.rows*r.columns; i++ {
		r.data[i] = value
	}
}

// Sets the data based on an existing array.
func (r *RectangularArrayFloat64) InitializeWithData(values []float64) error {
	// first check to see that it is the right length
	if len(values) == r.rows*r.columns {
		r.data = values
		return nil
	} else {
		return ArrayLengthError
	}
}

// A rectangular shaped array (matrix) of byte type. The array is not
// thread-safe. See ParallelRectangularArrayByte for a thread-safe implementation
type RectangularArrayByte struct {
	data          []byte
	rows, columns int
}

func NewRectangularArrayByte(rows, columns int) *RectangularArrayByte {
	r := RectangularArrayByte{rows: rows, columns: columns}
	r.data = make([]byte, rows*columns)
	return &r
}

// Returns the number of rows
func (r *RectangularArrayByte) GetRows() int {
	return r.rows
}

// Returns the number of columns
func (r *RectangularArrayByte) GetColumns() int {
	return r.columns
}

// Retrives an individual cell value in the matrix.
func (r *RectangularArrayByte) Value(row, column int) byte { //}, error) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		// the row and column are within the bounds of the matrix
		return r.data[row*r.columns+column] //, nil
	} //else {
	// the row and column are outside the bounds of the matrix
	return 0 //, NoDataError
	//}
}

// Sets an individual cell value in the matrix.
func (r *RectangularArrayByte) SetValue(row, column int, value byte) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		r.data[row*r.columns+column] = value
	} // else do nothing, the cell is outside the bounds of the matrix
}

func (r *RectangularArrayByte) GetRowData(row int) []byte {
	values := make([]byte, r.columns)
	for column := 0; column < r.columns; column++ {
		values[column] = r.data[row*r.columns+column]
	}
	return values
}

func (r *RectangularArrayByte) SetRowData(row int, values []byte) {
	if row >= 0 && row < r.rows {
		for column := 0; column < r.columns; column++ {
			r.data[row*r.columns+column] = values[column]
		}
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Increments an individual cell value in the matrix.
func (r *RectangularArrayByte) Increment(row, column int, values ...byte) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		if len(values) == 0 {
			r.data[row*r.columns+column]++
		} else {
			for _, num := range values {
				r.data[row*r.columns+column] += num
			}
		}
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Decrements an individual cell value in the matrix.
func (r *RectangularArrayByte) Decrement(row, column int, values ...byte) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		if len(values) == 0 {
			r.data[row*r.columns+column]--
		} else {
			for _, num := range values {
				r.data[row*r.columns+column] -= num
			}
		}
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Initializes all cells with a constant value.
func (r *RectangularArrayByte) InitializeWithConstant(value byte) {
	for i := 0; i < r.rows*r.columns; i++ {
		r.data[i] = value
	}
}

// Sets the data based on an existing array.
func (r *RectangularArrayByte) InitializeWithData(values []byte) error {
	// first check to see that it is the right length
	if len(values) == r.rows*r.columns {
		r.data = values
		return nil
	} else {
		return ArrayLengthError
	}
}

// A mutexByte is simply a thread-safe byte with accessors
type mutexByte struct {
	value byte
	sync.Mutex
}

func (this *mutexByte) get() byte {
	this.Lock()
	defer this.Unlock()
	return this.value
}

func (this *mutexByte) set(value byte) {
	this.Lock()
	defer this.Unlock()
	this.value = value
}

func (this *mutexByte) increment(value byte) {
	this.Lock()
	defer this.Unlock()
	this.value += value
	//	if len(values) == 0 {
	//		this.value++
	//	} else {
	//		for _, num := range values {
	//			this.value += num
	//		}
	//	}
}

func (this *mutexByte) decrement(value byte) {
	this.Lock()
	defer this.Unlock()
	this.value -= value
	//	if len(values) == 0 {
	//		this.value--
	//	} else {
	//		for _, num := range values {
	//			this.value -= num
	//		}
	//	}
}

func (this *mutexByte) incrementAndReturn(value byte) byte {
	this.Lock()
	defer this.Unlock()
	this.value += value
	return this.value
}

func (this *mutexByte) decrementAndReturn(value byte) byte {
	this.Lock()
	defer this.Unlock()
	this.value -= value
	return this.value
}

// A fine-grained concurrent rectangular shaped array (matrix) of byte type.
// The array is thread-safe and uses mutexes on each cell.
type ParallelRectangularArrayByte struct {
	data          []mutexByte
	rows, columns int
	sync.RWMutex
}

func NewParallelRectangularArrayByte(rows, columns int) *ParallelRectangularArrayByte {
	r := ParallelRectangularArrayByte{rows: rows, columns: columns}
	r.data = make([]mutexByte, rows*columns)
	//r.lock = &sync.Mutex{}
	return &r
}

// Returns the number of rows
func (r *ParallelRectangularArrayByte) GetRows() int {
	r.RLock()
	defer r.RUnlock()
	return r.rows
}

// Returns the number of columns
func (r *ParallelRectangularArrayByte) GetColumns() int {
	r.RLock()
	defer r.RUnlock()
	return r.columns
}

// Retrives an individual cell value in the matrix.
func (r *ParallelRectangularArrayByte) Value(row, column int) byte {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		// the row and column are within the bounds of the matrix
		return r.data[row*r.columns+column].get()
	}
	// the row and column are outside the bounds of the matrix
	return 0
}

// Sets an individual cell value in the matrix.
func (r *ParallelRectangularArrayByte) SetValue(row, column int, value byte) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		r.data[row*r.columns+column].set(value)
	} // else do nothing, the cell is outside the bounds of the matrix
}

func (r *ParallelRectangularArrayByte) GetRowData(row int) []byte {
	r.RLock()
	defer r.RUnlock()
	values := make([]byte, r.columns)
	for column := 0; column < r.columns; column++ {
		values[column] = r.data[row*r.columns+column].value
	}
	return values
}

func (r *ParallelRectangularArrayByte) SetRowData(row int, values []byte) {
	r.Lock()
	defer r.Unlock()
	if row >= 0 && row < r.rows {
		for column := 0; column < r.columns; column++ {
			r.data[row*r.columns+column].value = values[column]
		}
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Increments an individual cell value in the matrix.
func (r *ParallelRectangularArrayByte) Increment(row, column int, value byte) { // values ...byte) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		r.data[row*r.columns+column].increment(value)
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Decrements an individual cell value in the matrix.
func (r *ParallelRectangularArrayByte) Decrement(row, column int, value byte) { // values ...byte) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		r.data[row*r.columns+column].decrement(value)
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Increments an individual cell value in the matrix and return the value.
func (r *ParallelRectangularArrayByte) IncrementAndReturn(row, column int, value byte) byte { // values ...byte) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		return r.data[row*r.columns+column].incrementAndReturn(value)
	} // else do nothing, the cell is outside the bounds of the matrix
	return 0
}

// Decrements an individual cell value in the matrix and return the value.
func (r *ParallelRectangularArrayByte) DecrementAndReturn(row, column int, value byte) byte { // values ...byte) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		return r.data[row*r.columns+column].decrementAndReturn(value)
	} // else do nothing, the cell is outside the bounds of the matrix
	return 0
}

// Initializes all cells with a constant value.
func (r *ParallelRectangularArrayByte) InitializeWithConstant(value byte) {
	for i := 0; i < r.rows*r.columns; i++ {
		r.data[i].set(value)
	}
}

// Sets the data based on an existing array.
func (r *ParallelRectangularArrayByte) InitializeWithData(values []byte) error {
	// first check to see that it is the right length
	if len(values) == r.rows*r.columns {
		for i := 0; i < r.rows*r.columns; i++ {
			r.data[i].set(values[i])
		}
		return nil
	} else {
		return ArrayLengthError
	}
}

// A mutexFloat64 is simply a thread-safe float64 with accessors
type mutexFloat64 struct {
	value float64
	sync.Mutex
}

func (this *mutexFloat64) get() float64 {
	this.Lock()
	defer this.Unlock()
	return this.value
}

func (this *mutexFloat64) set(value float64) {
	this.Lock()
	defer this.Unlock()
	this.value = value
}

func (this *mutexFloat64) increment(value float64) {
	this.Lock()
	defer this.Unlock()
	this.value += value
	//	if len(values) == 0 {
	//		this.value++
	//	} else {
	//		for _, num := range values {
	//			this.value += num
	//		}
	//	}
}

func (this *mutexFloat64) decrement(value float64) {
	this.Lock()
	defer this.Unlock()
	this.value -= value
	//	if len(values) == 0 {
	//		this.value--
	//	} else {
	//		for _, num := range values {
	//			this.value -= num
	//		}
	//	}
}

func (this *mutexFloat64) incrementAndReturn(value float64) float64 {
	this.Lock()
	defer this.Unlock()
	this.value += value
	return this.value
}

func (this *mutexFloat64) decrementAndReturn(value float64) float64 {
	this.Lock()
	defer this.Unlock()
	this.value -= value
	return this.value
}

// A fine-grained concurrent rectangular shaped array (matrix) of float64 type.
// The array is thread-safe and uses mutexes on each cell.
type ParallelRectangularArrayFloat64 struct {
	data          []mutexFloat64
	rows, columns int
	nodata        float64
	sync.RWMutex
}

func NewParallelRectangularArrayFloat64(rows, columns int, nodata float64) *ParallelRectangularArrayFloat64 {
	r := ParallelRectangularArrayFloat64{rows: rows, columns: columns, nodata: nodata}
	r.data = make([]mutexFloat64, rows*columns)
	return &r
}

// Returns the number of rows
func (r *ParallelRectangularArrayFloat64) GetRows() int {
	r.RLock()
	defer r.RUnlock()
	return r.rows
}

// Returns the number of columns
func (r *ParallelRectangularArrayFloat64) GetColumns() int {
	r.RLock()
	defer r.RUnlock()
	return r.columns
}

// Returns the nodata value
func (r *ParallelRectangularArrayFloat64) GetNodata() float64 {
	r.RLock()
	defer r.RUnlock()
	return r.nodata
}

// Sets the nodata value
func (r *ParallelRectangularArrayFloat64) SetNodata(value float64) {
	r.Lock()
	defer r.Unlock()
	r.nodata = value
}

// Retrives an individual cell value in the matrix.
func (r *ParallelRectangularArrayFloat64) Value(row, column int) float64 {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		// the row and column are within the bounds of the matrix
		return r.data[row*r.columns+column].get()
	} else {
		// the row and column are outside the bounds of the matrix
		return r.nodata
	}
}

// Sets an individual cell value in the matrix.
func (r *ParallelRectangularArrayFloat64) SetValue(row, column int, value float64) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		r.data[row*r.columns+column].set(value)
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Returns an entire row of values.
func (r *ParallelRectangularArrayFloat64) GetRowData(row int) []float64 {
	r.RLock()
	defer r.RUnlock()
	values := make([]float64, r.columns)
	for column := 0; column < r.columns; column++ {
		values[column] = r.data[row*r.columns+column].value
	}
	return values
}

// Sets and entire row of values.
func (r *ParallelRectangularArrayFloat64) SetRowData(row int, values []float64) {
	r.Lock()
	defer r.Unlock()
	if row >= 0 && row < r.rows {
		for column := 0; column < r.columns; column++ {
			r.data[row*r.columns+column].value = values[column]
		}
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Increments an individual cell value in the matrix.
func (r *ParallelRectangularArrayFloat64) Increment(row, column int, value float64) { //values ...float64) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		r.data[row*r.columns+column].increment(value)
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Decrements an individual cell value in the matrix.
func (r *ParallelRectangularArrayFloat64) Decrement(row, column int, value float64) { // values ...float64) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		r.data[row*r.columns+column].decrement(value)
	} // else do nothing, the cell is outside the bounds of the matrix
}

// Increments an individual cell value in the matrix.
func (r *ParallelRectangularArrayFloat64) IncrementAndReturn(row, column int, value float64) float64 { //values ...float64) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		return r.data[row*r.columns+column].incrementAndReturn(value)
	} // else do nothing, the cell is outside the bounds of the matrix
	return r.nodata
}

// Decrements an individual cell value in the matrix.
func (r *ParallelRectangularArrayFloat64) DecrementAndReturn(row, column int, value float64) float64 { // values ...float64) {
	if column >= 0 && column < r.columns && row >= 0 && row < r.rows {
		return r.data[row*r.columns+column].decrementAndReturn(value)
	} // else do nothing, the cell is outside the bounds of the matrix
	return r.nodata
}

// Initializes all cells with a constant value.
func (r *ParallelRectangularArrayFloat64) InitializeWithConstant(value float64) {
	for i := 0; i < r.rows*r.columns; i++ {
		r.data[i].value = value
	}
}

// Sets the data based on an existing array.
func (r *ParallelRectangularArrayFloat64) InitializeWithData(values []float64) error {
	// first check to see that it is the right length
	if len(values) == r.rows*r.columns {
		for i := 0; i < r.rows*r.columns; i++ {
			r.data[i].value = values[i]
		}
		return nil
	} else {
		return ArrayLengthError
	}
}

// errors
var ArrayLengthError = errors.New("Incorrect array length: The specified data array must have rows * columns elements.")
var NoDataError = errors.New("There has been an attempt to access a cell beyond the grid edges.")
