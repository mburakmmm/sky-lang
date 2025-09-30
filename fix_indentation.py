#!/usr/bin/env python3
"""
Fix indentation in entry_rules_tester.py
"""

import re

def fix_indentation(content):
    # Fix brace blocks - content inside {} should be indented with 2 spaces
    lines = content.split('\n')
    fixed_lines = []
    in_brace_block = False
    brace_level = 0
    
    for line in lines:
        # Check if we're entering a brace block
        if '{' in line:
            in_brace_block = True
            brace_level = line.count('{') - line.count('}')
            fixed_lines.append(line)
            continue
        
        # Check if we're exiting a brace block
        if '}' in line:
            brace_level -= line.count('}')
            if brace_level <= 0:
                in_brace_block = False
            fixed_lines.append(line)
            continue
        
        # If we're inside a brace block, ensure proper indentation
        if in_brace_block and line.strip():
            # Remove existing indentation and add 2 spaces
            stripped = line.lstrip()
            if stripped:  # Only if line is not empty
                fixed_lines.append('  ' + stripped)
            else:
                fixed_lines.append(line)
        else:
            fixed_lines.append(line)
    
    return '\n'.join(fixed_lines)

# Read the file
with open('entry_rules_tester.py', 'r') as f:
    content = f.read()

# Fix indentation
fixed_content = fix_indentation(content)

# Write back
with open('entry_rules_tester.py', 'w') as f:
    f.write(fixed_content)

print("Indentation fixed!")
