# GoSpatial
##Description
GoSpatial is a simple command-line interface program for manipulating geospatial data, developed by [John Lindsay](http://www.uoguelph.ca/geography/faculty/lindsay-john "John Lindsay's homepage"). It is developed using the [Go](https://golang.org "Go programming language homepage") programming language and is compiled to native code. The project is experimental and is in its initial stage. GoSpatial is intended to provide additional analytical support for the [Whitebox Geospatial Analysis Tools](http://www.uoguelph.ca/~hydrogeo/Whitebox/ "Whitebox GAT homepage") open-source GIS software.

##Install
```
go get github.com/jblindsay/go-spatial
```

##Usage

###Getting help
To print a list of commands for GoSpatial, simply use the ```help``` command:

```
*************************
* Welcome to GoSpatial! *
*************************
Please enter a command: help
The following commands are recognized:
bench           Prints the current benchmarking mode
benchoff        Turns benchmarking mode off
benchon         Turns benchmarking mode on. Note: not all tools support this
clear           Clears the screen (also 'c', 'cls', or 'clr')
cwd             Changes the working directory (also 'cd' or 'dir'),
                 e.g. cwd /Users/john/
exit            Exits GoSpatial (also 'logout' or 'esc')
help            Prints a list of available commands (also 'h')
licence         Prints the licence
listtools       Lists all available tools
memprof         Outputs a memory usage profile
pwd             Prints the working directory (also 'dir')
rasterformats   Prints the supported raster formats
run             Runs a specified tool (also 'r'),
                 e.g. run toolname  or  run toolname "arg1;arg2;arg3;..."
toolargs        Prints the argument descriptions for a tool
toolhelp        Prints help documentation for a tool,
                 e.g. toolhelp BreachDepressions
version         Prints version information (also 'v')
Please enter a command:
```

The most common command that you will use is the ```run``` command.

###Working directories
To print the current working directory, use the ```pwd``` command:
```
Please enter a command: pwd
Working directory: /Users/johnlindsay/Documents/Data/
```

To change the working directory, use the ```cwd``` command:
```
Please enter a command: cwd /Users/johnlinsay/Documents/data
```

###Tools
To print a list of available tools, use the ```listtools``` command:
```
Please enter a command: listtools
The following tools 11 are available:
BreachDepressions    Removes depressions in DEMs using selective breaching
D8FlowAccumulation   Performs D8 flow accumulation on a DEM
DeviationFromMean    Calculates the deviation from mean
ElevationPercentile  Calculates the local elevation percentile for a DEM
FD8FlowAccum         Performs FD8 flow accumulation on a DEM
FillDepressions      Removes depressions in DEMs using filling
FillSmallNodataHoles Fills small nodata holes in a raster
MaxElevationDeviatio Calculates the maximum elevation deviation across a ran
PrintGeoTiffTags     Prints a GeoTiff's tags
Quantiles            Tranforms raster values into quantiles
Whitebox2GeoTiff     Converts Whitebox GAT raster to GeoTiff
```

To run a tool from the command-line interface, use the ```run``` command:

```
Please enter a command: run BreachDepressions
*********************
* breachdepressions *
*********************
Enter the DEM file name (incl. file extension):
```

When you run a tool from the command-line interface, you will be guided in terms of the input of the arguments required to run the tool. When you are prompted for an input file name (as in the above example), if the file resides in the current working directory, you can safely omit the directory from the file name. If you would like to print a list of arguments needed to run a particular plugin tool, with descriptions, use the ```toolargs``` command:

```
Please enter a command: toolargs breachdepressions
The following arguments are listed for 'breachdepressions':
InputDEM               string    The input DEM name with file extension
OutputFile             string    The output filename with file extension
MaxDepth               float64   The maximum breach channel depth (-1 to ignore)
MaxLength              int       The maximum length of a breach channel (-1 to ignore)
ConstrainedBreaching   bool      Use constrained breaching?
SubsequentFilling      bool      Perform post-breach filling?
Please enter a command:
```

This can be helpful when you want to execute a GoSpatial and run a tool by specifying flags and arguments. In the example below, after ```cd```ing to the directory containing the go-spatial executable file, it is possible to run a specific tool (```filldepressions```), providing the arguments for the ```-cwd```, ```-run```, and ```-args``` flags:

```
$ ./go-spatial -cwd="/Users/jlindsay/data/" -run="filldepressions" -args="my DEM.dep;outputDEM.tif;true"
*******************
* filldepressions *
*******************
Reading DEM data...
Filling DEM (2 of 2): 100%
Operation complete!
Elapsed time (excluding file I/O): 5.449522954s
Elapsed time (total): 6.087567077s
$
```

Sometimes you need to call a GoSpatial tool in an automated fashion, rather than using the GoSpatial command-line interface. Here is an example for how to call a GoSpatial tool from a Python script:

```python
#! /usr/bin/env python3
import subprocess

executablestr = "/Users/johnlindsay/Projects/whitebox/bin/gospatial"
workdir = "/Users/johnlindsay/Documents/Research/FastBreaching/data/"
toolname = "filldepressions"
args = "quebec DEM.dep;tmp11.dep;true"

a = [executablestr, "-cwd", workdir, "-run", toolname, "-args", args]

print("Setting up process...")
p = subprocess.Popen(a)
print("Running process...")
p.wait()
print("Done!")
```

##License
GoSpatial is distributed under the [MIT open-source license](./LICENSE).
