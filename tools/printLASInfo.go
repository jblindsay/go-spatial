// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Feb. 2015.

package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/jblindsay/go-spatial/geospatialfiles/lidar"
)

type PrintLASInfo struct {
	inputFile   string
	toolManager *PluginToolManager
}

func (this *PrintLASInfo) GetName() string {
	s := "PrintLASInfo"
	return getFormattedToolName(s)
}

func (this *PrintLASInfo) GetDescription() string {
	s := "Prints details of a LAS file"
	return getFormattedToolDescription(s)
}

func (this *PrintLASInfo) GetHelpDocumentation() string {
	ret := "This tool prints the metadata associated with a LAS file."
	return ret
}

func (this *PrintLASInfo) SetToolManager(tm *PluginToolManager) {
	this.toolManager = tm
}

func (this *PrintLASInfo) GetArgDescriptions() [][]string {
	numArgs := 1

	ret := make([][]string, numArgs)
	for i := range ret {
		ret[i] = make([]string, 3)
	}
	ret[0][0] = "InputFile"
	ret[0][1] = "string"
	ret[0][2] = "The input LAS file name"

	return ret
}

func (this *PrintLASInfo) ParseArguments(args []string) {
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

	this.Run()
}

func (this *PrintLASInfo) CollectArguments() {
	consolereader := bufio.NewReader(os.Stdin)

	// get the input file name
	print("Enter the  file name (incl. file extension): ")
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

	this.Run()
}

func (this *PrintLASInfo) Run() {

	input, err := lidar.CreateFromFile(this.inputFile)
	if err != nil {
		println(err.Error())
	}
	defer input.Close()

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("File Name: %v\n", input.GetFileName()))
	day, month := convertYearday(int(input.Header.FileCreationDay), int(input.Header.FileCreationYear))
	buffer.WriteString(fmt.Sprintf("Creation Date: %v %v, %v\n", day, month, input.Header.FileCreationYear))
	buffer.WriteString(fmt.Sprintf("Generating software: %v\n", input.Header.GeneratingSoftware))
	buffer.WriteString(fmt.Sprintf("LAS version: %v.%v\n", input.Header.VersionMajor, input.Header.VersionMajor))
	buffer.WriteString(fmt.Sprintf("Number of Points: %v\n", input.Header.NumberPoints))
	buffer.WriteString(fmt.Sprintf("Point record length: %v\n", input.Header.PointRecordLength))
	buffer.WriteString(fmt.Sprintf("Point record format: %v\n", input.Header.PointFormatID))

	buffer.WriteString(fmt.Sprintf("Point record format: %v\n", input.Header.String()))

	println(buffer.String())
}

func convertYearday(yday int, year int) (int, string) {
	var months = [...]string{
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"October",
		"November",
		"December",
	}

	var lastDayOfMonth = [...]int{
		31,
		59,
		90,
		120,
		151,
		181,
		212,
		243,
		273,
		304,
		334,
		365,
	}

	var month int
	day := yday
	if isLeapYear(year) {
		// Leap year
		switch {
		case day > 31+29-1:
			// After leap day; pretend it wasn't there.
			day--
		case day == 31+29-1:
			// Leap day.
			month = 1
			day = 29
			return day, months[month]
		}
	}

	month = 0
	for m := range lastDayOfMonth {
		//for m := 0; m < 12; m++ {
		if day > lastDayOfMonth[m] {
			month++
		} else {
			break
		}
	}
	if month > 0 {
		day -= lastDayOfMonth[month-1]
	}

	return day, months[month]
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
