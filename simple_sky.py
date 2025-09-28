#!/usr/bin/env python3
"""
Sky Language - Basit Demo Interpreter
Sky dilinin temel özelliklerini gösterir
"""

import re
import sys

class SkyInterpreter:
    def __init__(self):
        self.variables = {}
        
    def interpret(self, code):
        lines = code.strip().split('\n')
        for line in lines:
            line = line.strip()
            if not line or line.startswith('#'):
                continue
            self.execute_line(line)
    
    def execute_line(self, line):
        # print("Merhaba Sky!")
        if line.startswith('print(') and line.endswith(')'):
            content = line[6:-1].strip()
            result = self.evaluate_expression(content)
            print(result)
        
        # int sayı = 42
        elif ' = ' in line and not line.startswith('print'):
            parts = line.split(' = ')
            if len(parts) == 2:
                var_part = parts[0].strip()
                value_part = parts[1].strip()
                
                # Tip kontrolü
                if var_part.startswith('int '):
                    var_name = var_part[4:].strip()
                    try:
                        self.variables[var_name] = int(value_part)
                    except ValueError:
                        print(f"Hata: {value_part} bir sayı değil")
                elif var_part.startswith('string '):
                    var_name = var_part[7:].strip()
                    if value_part.startswith('"') and value_part.endswith('"'):
                        self.variables[var_name] = value_part[1:-1]
                    else:
                        print(f"Hata: String değer tırnak içinde olmalı")
                elif var_part.startswith('var '):
                    var_name = var_part[4:].strip()
                    # Otomatik tip belirleme
                    if value_part.startswith('"') and value_part.endswith('"'):
                        self.variables[var_name] = value_part[1:-1]
                    else:
                        try:
                            self.variables[var_name] = int(value_part)
                        except ValueError:
                            try:
                                self.variables[var_name] = float(value_part)
                            except ValueError:
                                self.variables[var_name] = value_part
    
    def evaluate_expression(self, expr):
        """Basit expression evaluation"""
        expr = expr.strip()
        
        # String literal
        if expr.startswith('"') and expr.endswith('"'):
            return expr[1:-1]
        
        # Variable
        if expr in self.variables:
            return self.variables[expr]
        
        # String concatenation: "hello" + var
        if ' + ' in expr:
            parts = expr.split(' + ')
            result = ""
            for part in parts:
                part = part.strip()
                if part.startswith('"') and part.endswith('"'):
                    result += part[1:-1]
                elif part in self.variables:
                    result += str(self.variables[part])
                else:
                    # Try to evaluate as expression
                    evaluated = self.evaluate_expression(part)
                    if isinstance(evaluated, (int, float, str)):
                        result += str(evaluated)
                    else:
                        result += part
            return result
        
        # Simple arithmetic
        if ' + ' in expr and not '"' in expr:
            parts = expr.split(' + ')
            try:
                total = 0
                for part in parts:
                    part = part.strip()
                    if part in self.variables:
                        if isinstance(self.variables[part], (int, float)):
                            total += self.variables[part]
                        else:
                            return expr  # Non-numeric variable
                    else:
                        total += int(part)
                return int(total) if total == int(total) else total
            except:
                return expr
        
        # Number literal
        try:
            return int(expr)
        except ValueError:
            try:
                return float(expr)
            except ValueError:
                return expr

def main():
    if len(sys.argv) > 1:
        # Dosyadan oku
        with open(sys.argv[1], 'r', encoding='utf-8') as f:
            code = f.read()
    else:
        # REPL modu
        print("Sky Language Demo - Çıkmak için 'quit' yazın")
        interpreter = SkyInterpreter()
        while True:
            try:
                line = input("sky> ")
                if line.strip() == 'quit':
                    break
                interpreter.execute_line(line.strip())
            except KeyboardInterrupt:
                break
        return
    
    interpreter = SkyInterpreter()
    interpreter.interpret(code)

if __name__ == "__main__":
    main()
