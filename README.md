# JSON Redaction / Replacement
This program reads file(s) containing JSON records, and redacts or replaces the
private / client information in each based on some predefined parameters.

All three flags must be specified.

Usage:

	./json_replace [flags]

Flags:

	-i input_path
		Set the path to the input file or directory. Path to a file must be a json file.

	-o output_path
		Set the path to the out file or directory. Path to a file must be a json file.

	-c config_path
		Set the path to the config file. The file must be a json file.
