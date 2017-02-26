#!/usr/bin/env python
import sys
import gospatial as gs

def main():
    try:
        # Get the gospatial version number
        print(gs.version())

        # List all available tools in gospatial
        print(gs.list_tools())

        # Prints the gospatial help...a listing of available commands
        print(gs.help())

        # Print the help documentation for the Aspect tool
        print(gs.tool_help("Aspect"))

        # Prints the arguments used for running the FillDepressions tool
        print(gs.tool_args("FillDepressions"))

        # Sets the working directory. If /the working dir is set, you don't
        # need to specify complete file names (with paths) to tools that you run.
        gs.set_working_dir("/Users/johnlindsay/Documents/data/JayStateForest")

        # Run the Aspect tool, specifying the arguments.
        name = "Aspect"
        args = [
            "DEM.dep",
            "aspect.dep"
        ]

        # Run the tool and check the return value
        ret = gs.run_tool(name, args, callback)
        if ret != 0:
            print("ERROR: return value={}".format(ret))

    except:
        print("Unexpected error:", sys.exc_info()[0])
        raise

# Create a custom callback to process the text coming out of the tool.
# If a callback is not provided, it will simply print the output stream.
# A provided callback allows for custom processing of the output stream.
def callback(s):
    try:
        # print("{}".format(s))
        if "%" in s:
            str_array = s.split(" ")
            label = s.replace(str_array[len(str_array)-1], "")
            progress = int(str_array[len(str_array)-1].replace("%", "").strip())
            print("Progress: {}%".format(progress))
        else:
            if "error" in s.lower():
                print("ERROR: {}".format(s))
            else:
                print("{}".format(s))
    except:
        print(s)

main()
