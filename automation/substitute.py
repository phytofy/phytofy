#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
This script replaces a string in output file with the contents
of the input file.
"""

import sys


if __name__ == '__main__':
    if len(sys.argv) != 4:
        sys.exit(1)
    else:
        input_file = sys.argv[1]
        output_file = sys.argv[2]
        placeholder = sys.argv[3]
        with open(input_file, 'r') as handle:
            input_contents = handle.read()
        with open(output_file, 'r') as handle:
            output_contents = handle.read()
        output_contents = output_contents.replace(placeholder, input_contents)
        with open(output_file, 'w') as handle:
            handle.write(output_contents)
