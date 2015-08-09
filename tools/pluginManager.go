// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// This file was originally created by John Lindsay<jlindsay@uoguelph.ca>,
// Feb. 2015.

package tools

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

var println = fmt.Println
var printf = fmt.Printf
var print = fmt.Print
var pathSep string = string(os.PathSeparator)

type PluginToolManager struct {
	workingDirectory string
	mapOfPluginTools map[string]PluginTool
	BenchMode        bool
}

func (ptm *PluginToolManager) InitializeTools() {
	// each new tool needs a two-line entry below
	ptm.mapOfPluginTools = make(map[string]PluginTool)

	hillshade := new(Hillshade)
	ptm.mapOfPluginTools[strings.ToLower(hillshade.GetName())] = hillshade

	aspect := new(Aspect)
	ptm.mapOfPluginTools[strings.ToLower(aspect.GetName())] = aspect

	slope := new(Slope)
	ptm.mapOfPluginTools[strings.ToLower(slope.GetName())] = slope

	wb2gt := new(Whitebox2GeoTiff)
	ptm.mapOfPluginTools[strings.ToLower(wb2gt.GetName())] = wb2gt

	bd := new(BreachDepressions)
	ptm.mapOfPluginTools[strings.ToLower(bd.GetName())] = bd

	d8fa := new(D8FlowAccumulation)
	ptm.mapOfPluginTools[strings.ToLower(d8fa.GetName())] = d8fa

	fd8fa := new(FD8FlowAccum)
	ptm.mapOfPluginTools[strings.ToLower(fd8fa.GetName())] = fd8fa

	fd := new(FillDepressions)
	ptm.mapOfPluginTools[strings.ToLower(fd.GetName())] = fd

	pgtt := new(PrintGeoTiffTags)
	ptm.mapOfPluginTools[strings.ToLower(pgtt.GetName())] = pgtt

	pli := new(PrintLASInfo)
	ptm.mapOfPluginTools[strings.ToLower(pli.GetName())] = pli

	dfm := new(DeviationFromMean)
	ptm.mapOfPluginTools[strings.ToLower(dfm.GetName())] = dfm

	med := new(MaximumElevationDeviation)
	ptm.mapOfPluginTools[strings.ToLower(med.GetName())] = med

	ep := new(ElevationPercentile)
	ptm.mapOfPluginTools[strings.ToLower(ep.GetName())] = ep

	q := new(Quantiles)
	ptm.mapOfPluginTools[strings.ToLower(q.GetName())] = q

	fsnh := new(FillSmallNodataHoles)
	ptm.mapOfPluginTools[strings.ToLower(fsnh.GetName())] = fsnh

}

func (ptm *PluginToolManager) GetListOfTools() []PluginTool {
	ret := make([]PluginTool, len(ptm.mapOfPluginTools)) //ptm.listOfPluginTools
	i := 0
	for _, val := range ptm.mapOfPluginTools {
		ret[i] = val
		i++
	}
	return ret
}

func (ptm *PluginToolManager) Run(toolName string) error {
	toolName = strings.ToLower(getFormattedToolName(toolName))
	if tool, ok := ptm.mapOfPluginTools[toolName]; ok {
		//do something here
		println(GetHeaderText(toolName))
		tool.SetToolManager(ptm)
		tool.CollectArguments()
		runtime.GC()
		return nil
	} else {
		return errors.New("Unrecognized tool name. Type 'listtools' for a list of available tools.\n")
	}
}

func (ptm *PluginToolManager) RunWithArguments(toolName string, args []string) error {
	toolName = strings.ToLower(getFormattedToolName(toolName))
	if tool, ok := ptm.mapOfPluginTools[toolName]; ok {
		//do something here
		println(GetHeaderText(toolName))
		tool.SetToolManager(ptm)
		tool.ParseArguments(args)
		runtime.GC()
		return nil
	} else {
		return errors.New("Unrecognized tool name. Type 'listtools' for a list of available tools.\n")
	}
}

func (ptm *PluginToolManager) GetToolArgDescriptions(toolName string) ([]string, error) {
	trailingSpaces := func(s string, maxLen int) string {
		strLen := len(s)
		sepSpace := maxLen - strLen
		sepStr := " "
		for i := 0; i < sepSpace; i++ {
			sepStr += " "
		}
		return s + sepStr
	}

	toolName = strings.ToLower(getFormattedToolName(toolName))
	if tool, ok := ptm.mapOfPluginTools[toolName]; ok {
		descEntries := tool.GetArgDescriptions()
		lenToolName := 0
		lenDataType := 0
		for _, val := range descEntries {
			if len(val[0]) > lenToolName {
				lenToolName = len(val[0])
			}
			if len(val[1]) > lenDataType {
				lenDataType = len(val[1])
			}
		}

		lenToolName += 2
		lenDataType += 2

		ret := make([]string, len(descEntries))
		for i, val := range descEntries {
			ret[i] = trailingSpaces(val[0], lenToolName) + trailingSpaces(val[1], lenDataType) + val[2]
		}
		return ret, nil
	} else {
		return nil, errors.New("Unrecognized tool name. Type 'listtools' for a list of available tools.\n")
	}
}

func (ptm *PluginToolManager) SetWorkingDirectory(wd string) {
	if !strings.HasSuffix(wd, pathSep) {
		wd += pathSep
	}
	ptm.workingDirectory = wd
}

type PluginTool interface {
	GetName() string
	GetDescription() string
	GetHelpDocumentation() string
	CollectArguments()
	ParseArguments([]string)
	GetArgDescriptions() [][]string
	SetToolManager(*PluginToolManager)
}

type PluginToolList []PluginTool

func (ptl PluginToolList) Len() int { return len(ptl) }

func (ptl PluginToolList) Less(i, j int) bool {
	return ptl[i].GetName() < ptl[j].GetName()
}

func (ptl PluginToolList) Swap(i, j int) {
	ptl[i], ptl[j] = ptl[j], ptl[i]
}

func GetHeaderText(str string) string {
	ret := ""
	for i := 0; i < len(str)+4; i++ {
		ret += "*"
	}
	ret += "\n* "
	ret += str
	ret += " *\n"
	for i := 0; i < len(str)+4; i++ {
		ret += "*"
	}
	return ret
}

var maxToolNameLength = 20

func getFormattedToolName(s string) string {
	l := len(s)
	if l > maxToolNameLength {
		l = maxToolNameLength
	}
	return strings.TrimSpace(s[:l])
}

var maxToolDescriptionLength = 55

func getFormattedToolDescription(s string) string {
	l := len(s)
	if l > maxToolDescriptionLength {
		l = maxToolDescriptionLength
	}
	return strings.TrimSpace(s[:l])
}

func (ptm *PluginToolManager) GetToolHelp(toolName string) (string, error) {
	toolName = strings.ToLower(getFormattedToolName(toolName))
	if tool, ok := ptm.mapOfPluginTools[toolName]; ok {
		//showToolHelp(tool)
		return tool.GetHelpDocumentation(), nil
	} else {
		return "", errors.New("Unrecognized tool name. Type 'listtools' for a list of available tools.\n")
	}
}
