# go-spatial
GoSpatial is a simple command-line interface program for manipulating geospatial data

###Calling GoSpatial tools
Sometimes you need to call a GoSpatial tool in an automated fashion, rather than using the GoSpatial command-line interface. Here is an example for how to call a GoSpatial tool from a Python script:

```python
#! /usr/bin/env python3
#from subprocess import call
import subprocess

executablestr = "/Users/johnlindsay/Projects/whitebox/bin/gospatial"
workdir = "/Users/johnlindsay/Documents/Research/FastBreaching/data/"
toolname = "filldepressions"
args = "quebec DEM.dep;tmp11.dep;true"

a = [executablestr, "-cwd", workdir, "-run", toolname, "-args", args]

#call(a)
print("Setting up process...")
p = subprocess.Popen(a)
print("Running process...")
p.wait()
print("Done!")
```
