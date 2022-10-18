# JSON Redaction / Replacement
This program reads file(s) containing JSON records, and redacts or replaces the
private / client information in each based on some predefined parameters.

-i, -o, and -c flags must be specified.
-l and -r flags are optional.

Reading multiple JSON objects line-by-line is supported by specifying -l flag.
Note that a single JSON object in multiple lines is not supported if line-by-line mode is enabled.

The program is running in concurrency by default.

Usage:

	./json_replace [flags]

Flags:

	-i input_path
		Set the path to the input file or directory. Path to a file must be a json file.

	-o output_path
		Set the path to the out file or directory. Path to a file must be a json file.

	-c config_path
		Set the path to the config file. The file must be a json file.

	-l
		Read multiple json objects line by line. Default: false

	-r [number of routines]
		Set the maximum number of routines running simultaneously. Default: 10