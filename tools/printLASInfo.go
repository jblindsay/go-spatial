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

	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Creation Year: %v\n", input.Header.FileCreationYear))
	buffer.WriteString(fmt.Sprintf("Number of Points: %v\n", input.Header.NumberPoints))
	buffer.WriteString(fmt.Sprintf("Generating software: %v\n", input.Header.GeneratingSoftware))

	println(buffer.String())
}
