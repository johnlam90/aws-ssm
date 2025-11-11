#!/usr/bin/env python3
"""
Fix gosec SARIF output by removing invalid 'fixes' field.

The gosec tool generates SARIF files with an invalid 'fixes' field that doesn't
conform to the SARIF specification, causing upload failures to GitHub Code Scanning.
This script removes those invalid fields.
"""

import json
import sys
import shutil
from pathlib import Path


def fix_gosec_sarif(input_file: str, output_file: str) -> int:
    """
    Remove invalid 'fixes' fields from gosec SARIF output.
    
    Args:
        input_file: Path to the raw gosec SARIF file
        output_file: Path to write the fixed SARIF file
        
    Returns:
        Number of fixes fields removed
    """
    try:
        with open(input_file, 'r') as f:
            data = json.load(f)
        
        fixed_count = 0
        for run in data.get('runs', []):
            for result in run.get('results', []):
                if 'fixes' in result:
                    del result['fixes']
                    fixed_count += 1
        
        with open(output_file, 'w') as f:
            json.dump(data, f, indent=2)
        
        return fixed_count
        
    except Exception as e:
        print(f"Error fixing SARIF: {e}", file=sys.stderr)
        # Try to copy the raw file as fallback
        try:
            shutil.copy(input_file, output_file)
            print(f"Copied raw file to {output_file} (may fail upload)", file=sys.stderr)
        except Exception as copy_error:
            print(f"Failed to copy file: {copy_error}", file=sys.stderr)
        raise


def main():
    input_file = 'gosec-raw.sarif'
    output_file = 'gosec.sarif'
    
    if not Path(input_file).exists():
        print(f"Input file {input_file} not found", file=sys.stderr)
        sys.exit(1)
    
    try:
        fixed_count = fix_gosec_sarif(input_file, output_file)
        print(f"✅ Fixed gosec.sarif format (removed {fixed_count} invalid fixes fields)")
        sys.exit(0)
    except Exception as e:
        print(f"❌ Failed to fix gosec.sarif: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()

