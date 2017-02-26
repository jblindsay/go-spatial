package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
	"github.com/jblindsay/go-spatial/tools"
)

var version = "0.1.1"

// var githash = "0000"
var buildstamp = "no build stamp provided"

var println = fmt.Println
var printf = fmt.Printf
var print = fmt.Print
var printerr = func(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
}
var printErrString = func(s string) {
	fmt.Fprintln(os.Stderr, s)
}
var pathSep = string(os.PathSeparator)
var commandArgs []string
var carryon bool
var workingdir string
var err error
var toolManager tools.PluginToolManager

//var flagCpuprofile string

func main() {
	//flag.StringVar(&flagCpuprofile, "cpuprofile", "", "output a cpuprofile to...")
	var runTool string
	flag.StringVar(&runTool, "run", "", "Run a particular tool")
	var toolArgs string
	flag.StringVar(&toolArgs, "args", "", "Specify tool arguments, delimited by commas or semicolons")
	var cwd string
	flag.StringVar(&cwd, "cwd", "", "Change the working directory")
	var listTools = false
	flag.BoolVar(&listTools, "listtools", false, "Lists all available tools")
	var toolHelp string
	flag.StringVar(&toolHelp, "toolhelp", "", "Prints help documentation for a tool")
	var toolArgsStr string
	flag.StringVar(&toolArgsStr, "toolargs", "", "Prints details about the arguments for a tool")
	var helpArg = false
	flag.BoolVar(&helpArg, "help", false, "Help")
	// var ldflags string
	// flag.StringVar(&ldflags, "ldflags", "", "ldflags")
	var versionFlag = false
	flag.BoolVar(&versionFlag, "version", false, "Version number")
	flag.Parse()

	if strings.Contains(cwd, "\"") {
		cwd = strings.Replace(cwd, "\"", "", -1)
	}

	if strings.Contains(runTool, "\"") {
		runTool = strings.Replace(runTool, "\"", "", -1)
	}

	//	if flagCpuprofile != "" {
	//		f, err := os.Create(flagCpuprofile)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		pprof.StartCPUProfile(f)
	//		//defer pprof.StopCPUProfile()
	//	}

	//args := os.Args[1:]
	if listTools {
		if cmd, ok := commandMap["listtools"]; ok {
			cmd()
		} else {
			printerr(fmt.Errorf("unrecognized command '%s', type 'help' for details...", commandArgs[0]))
		}
	} else if versionFlag {
		if cmd, ok := commandMap["version"]; ok {
			cmd()
		} else {
			printerr(fmt.Errorf("unrecognized command '%s', type 'help' for details...", commandArgs[0]))
		}
	} else if helpArg {
		if cmd, ok := commandMap["help"]; ok {
			cmd()
		} else {
			printerr(fmt.Errorf("unrecognized command '%s', type 'help' for details...", commandArgs[0]))
		}
	} else if toolHelp != "" {
		commandArgs = []string{"toolhelp", toolHelp}
		if cmd, ok := commandMap["toolhelp"]; ok {
			cmd()
		} else {
			printerr(fmt.Errorf("Unrecognized command '%s', type 'help' for details...", commandArgs[0]))
		}
	} else if toolArgsStr != "" {
		commandArgs = []string{"toolargs", toolArgsStr}
		if cmd, ok := commandMap["toolargs"]; ok {
			cmd()
		} else {
			printerr(fmt.Errorf("Unrecognized command '%s', type 'help' for details...", commandArgs[0]))
		}
	} else if runTool != "" {
		//		var runTool string
		//		flag.StringVar(&runTool, "run", "", "Run a particular tool")
		//		var toolArgs string
		//		flag.StringVar(&toolArgs, "args", "", "Specify tool arguments, delimited by commas or semicolons")
		//		var cwd string
		//		flag.StringVar(&cwd, "cwd", "", "Change the working directory")
		//		flag.Parse()
		// fmt.Println(runTool)
		if len(strings.TrimSpace(cwd)) > 0 {
			changeWorkingDirectory(cwd)
		}
		toolArgs = strings.Replace(toolArgs, "%s", " ", -1)
		argsArray := []string{}
		if len(toolArgs) > 0 {
			// parse the args
			f := func(c rune) bool {
				return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '.' && c != os.PathSeparator && c != ' ' && c != '-' && c != '_'
			}
			argsArray = strings.FieldsFunc(toolArgs, f)
		}
		if len(strings.TrimSpace(runTool)) > 0 {
			if err = toolManager.RunWithArguments(strings.TrimSpace(runTool), argsArray); err != nil {
				printerr(err)
				//printerr(fmt.Errorf("Unrecognized tool name '%s;. Type 'listtools' for a list of available tools.", commandArgs[1]))
			}
		}
	} else {
		// run it in command line mode
		println(getHeaderText("Welcome to GoSpatial"))
		consolereader := bufio.NewReader(os.Stdin)
		carryon = true

		// This is the main command loop.
		println("Type 'help' to review available commands and 'exit' to log out.")
		for carryon {
			print("Please enter a command: ")
			commandStr, err := consolereader.ReadString('\n')
			if err != nil {
				printerr(err)
				os.Exit(0)
			}
			commandStr = strings.TrimSpace(commandStr)
			if len(commandStr) > 0 {
				commandArgs = strings.Fields(commandStr)
				if cmd, ok := commandMap[strings.ToLower(commandArgs[0])]; ok {
					cmd()
				} else {
					printerr(fmt.Errorf("unrecognized command '%s', type 'help' for details...", commandArgs[0]))
				}
			} else {
				printErrString("Empty command, type 'help' for details...")
			}
		}
	}
}

var helpMap map[string][]string
var commandMap map[string]func()

func init() {
	toolManager = tools.PluginToolManager{}
	toolManager.InitializeTools()

	// set the current working directory
	if workingdir, err = os.Getwd(); err != nil {
		println("Error")
	}

	helpMap = make(map[string][]string)
	helpMap["clear"] = []string{"Clears the screen (also 'c', 'cls', or 'clr')"}
	helpMap["help"] = []string{"Prints a list of available commands (also 'h')"}
	helpMap["exit"] = []string{"Exits GoSpatial (also 'logout' or 'esc')"}
	helpMap["rasterformats"] = []string{"Prints the supported raster formats"}
	helpMap["version"] = []string{"Prints version information (also 'v')"}
	helpMap["cwd"] = []string{"Changes the working directory (also 'cd' or 'dir'),", " e.g. cwd /Users/john/"}
	helpMap["pwd"] = []string{"Prints the working directory (also 'dir')"}
	helpMap["run"] = []string{"Runs a specified tool (also 'r'),",
		" e.g. run toolname  or  run toolname \"arg1;arg2;arg3;...\""}
	helpMap["listtools"] = []string{"Lists all available tools"}
	helpMap["licence"] = []string{"Prints the licence"}
	helpMap["toolargs"] = []string{"Prints the argument descriptions for a tool"}
	helpMap["memprof"] = []string{"Outputs a memory usage profile"}
	helpMap["toolhelp"] = []string{"Prints help documentation for a tool,", " e.g. toolhelp BreachDepressions"}
	helpMap["benchon"] = []string{"Turns benchmarking mode on. Note: not all tools support this"}
	helpMap["benchoff"] = []string{"Turns benchmarking mode off"}
	helpMap["bench"] = []string{"Prints the current benchmarking mode"}

	commandMap = make(map[string]func())
	commandMap["benchon"] = func() {
		toolManager.BenchMode = true
	}
	commandMap["benchoff"] = func() {
		toolManager.BenchMode = false
	}
	commandMap["bench"] = func() {
		if toolManager.BenchMode {
			println("Benchmark Mode = on")
		} else {
			println("Benchmark Mode = off")
		}
	}
	commandMap["toolhelp"] = func() {
		if len(commandArgs) > 1 {
			s, err := toolManager.GetToolHelp(commandArgs[1])
			if err != nil {
				printf("Unrecognized tool name '%s'. Type 'listtools' for a list of available tools.\n", commandArgs[1])
			} else {
				println(s)
			}
		} else {
			println("Tool name not specified, e.g. toolhelp BreachDepressions")
		}
	}
	commandMap["clear"] = func() {
		callClear()
	}
	commandMap["clr"] = commandMap["clear"]
	commandMap["cls"] = commandMap["clear"]
	commandMap["c"] = commandMap["clear"]
	commandMap["help"] = func() {
		// first sort the commands alphabetically
		keys := make([]string, 0, len(helpMap))
		for key := range helpMap {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		println("The following commands are recognized:")
		for _, key := range keys {
			val := helpMap[key]
			println(trailingSpaces(key, 15) + val[0])
			if len(val) > 1 {
				for i := 1; i < len(val); i++ {
					println(trailingSpaces("", 15) + val[i])
				}
			}
		}
	}
	commandMap["h"] = commandMap["help"]
	commandMap["exit"] = func() {
		carryon = false
		println("Goodbye for now")
		//		if flagCpuprofile != "" {
		//			pprof.StopCPUProfile()
		//		}
		os.Exit(0)
	}
	commandMap["logout"] = commandMap["exit"]
	commandMap["esc"] = commandMap["exit"]
	commandMap["run"] = func() {
		if len(commandArgs) == 2 {
			if err = toolManager.Run(commandArgs[1]); err != nil {
				printf("Unrecognized tool name '%s'. Type 'listtools' for a list of available tools.\n", commandArgs[1])
			}
		} else if len(commandArgs) > 2 { // there are specified arguments
			s := ""
			for i := 2; i < len(commandArgs); i++ {
				s += " " + commandArgs[i]
			}
			s = strings.TrimSpace(s)
			// parse the args
			f := func(c rune) bool {
				return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '.' && c != os.PathSeparator && c != ' ' && c != '-'
			}
			argsArray := strings.FieldsFunc(s, f)

			if err = toolManager.RunWithArguments(strings.TrimSpace(commandArgs[1]), argsArray); err != nil {
				printf("Unrecognized tool name '%s'. Type 'listtools' for a list of available tools.\n", commandArgs[1])
			}
		} else {
			println("Tool name not specified, e.g. run BreachDepressions")
		}
	}
	commandMap["r"] = commandMap["run"]
	commandMap["rasterformats"] = func() {
		// first sort the commands alphabetically
		m := raster.GetMapOfFormatsAndExtensions()
		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		println("The following raster formats are supported for reading/writing:")
		for _, key := range keys {
			if !strings.Contains(strings.ToLower(key), "unknown") {
				val := m[key]
				println(trailingSpaces(key, 20), val)
			}
		}
	}
	commandMap["version"] = func() {
		printf("GoSpatial version %s.%s\n", version, buildstamp) //releaseDate.Format(layout))
	}
	commandMap["v"] = commandMap["version"]
	commandMap["pwd"] = func() {
		println("Working directory:", workingdir)
	}
	commandMap["cwd"] = func() {
		if len(commandArgs) > 1 {
			if len(commandArgs) == 2 {
				changeWorkingDirectory(commandArgs[1])
			} else {
				// This occurs when the directory has spaces. Paste the pieces of the directory back together.
				str := ""
				for i := 1; i < len(commandArgs); i++ {
					str += commandArgs[i] + " "
				}
				str = strings.TrimSpace(str)
				changeWorkingDirectory(str)
			}
		} else {
			println("A directory must be specified after the 'cwd' keyword.")
		}
	}
	commandMap["cd"] = commandMap["cwd"]
	commandMap["dir"] = func() {
		if len(commandArgs) > 1 {
			commandMap["cwd"]()
		} else {
			commandMap["pwd"]()
		}
	}

	commandMap["listtools"] = func() {
		pt := toolManager.GetListOfTools()
		plugs := make([]string, 0, len(pt))
		for _, value := range pt {
			plugs = append(plugs, trailingSpaces(value.GetName(), 20)+value.GetDescription())
		}
		sort.Strings(plugs)
		printf("The following %v tools are available:\n", len(pt))
		for _, value := range plugs {
			println(value)
		}
	}
	commandMap["licence"] = func() {
		println(licenceText)
	}
	commandMap["toolargs"] = func() {
		if len(commandArgs) > 1 {
			argDescriptions, err := toolManager.GetToolArgDescriptions(commandArgs[1])
			if err != nil {
				printf("Unrecognized tool name '%s'. Type 'listtools' for a list of available tools.\n", commandArgs[1])
			} else {
				printf("The following arguments are listed for '%s':\n", commandArgs[1])
				for _, val := range argDescriptions {
					println(val)
				}
			}
		} else {
			println("Tool name not specified, e.g. toolargs FastBreach")
		}
	}
	commandMap["memprof"] = func() {
		m := new(runtime.MemStats)
		runtime.ReadMemStats(m)
		println("Memory allocated and in current use =", (float64(m.Alloc) / 1000000.0), "MB")
		println("Total memory allocated =", (float64(m.TotalAlloc) / 1000000.0), "MB")
		println("Heap allocated and still in use =", (float64(m.HeapAlloc) / 1000000.0), "MB")
		println("Stack memory used by stack allocator =", (float64(m.StackInuse) / 1000000.0), "MB")
	}
}

var clear map[string]func() //create a map for storing clear funcs

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cls") //Windows example it is untested, but I think its working
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func callClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
		println(getHeaderText("Welcome to GoSpatial!"))
	} else { //unsupported platform
		println("Clearing the screen is unsupported for your platform.")
	}
}

var changeWorkingDirectory = func(wd string) {
	// see if the string is an existing directory
	if _, err := os.Stat(wd); err != nil {
		if os.IsNotExist(err) {
			// see if appending this directory to the working directory works
			if strings.HasPrefix(wd, pathSep) || strings.HasPrefix(wd, "."+pathSep) {
				if strings.HasPrefix(wd, "."+pathSep) {
					// remove the dot
					wd = wd[1:]
				}
				s := workingdir
				if strings.HasSuffix(s, pathSep) {
					s = s[:len(s)-1]
				}
				s += wd
				if _, err := os.Stat(s); err != nil {
					if os.IsNotExist(err) {
						println("Directory does not exist.")
					} else {
						println(err)
					}
				} else {
					workingdir = s
					toolManager.SetWorkingDirectory(s)
				}
			}
		} else {
			println(err)
		}
	} else {
		workingdir = wd
		toolManager.SetWorkingDirectory(wd)
	}
}
var licenceText = `Copyright (c) 2015 The GoSpatial Authors
Lead Developer: John Lindsay, PhD (jlindsay@uoguelph.ca),
The University of Guelph, Canada

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`

func getHeaderText(str string) string {
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

var trailingSpaces = func(s string, maxLen int) string {
	strLen := len(s)
	sepSpace := maxLen - strLen
	sepStr := " "
	for i := 0; i < sepSpace; i++ {
		sepStr += " "
	}
	return s + sepStr
}
