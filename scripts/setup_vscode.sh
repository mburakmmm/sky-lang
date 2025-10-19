#!/bin/bash
# VS Code setup script for SKY language

echo "🎨 Setting up VS Code for SKY..."
echo ""

# Create .vscode directory if it doesn't exist
mkdir -p .vscode

# Create settings.json
cat > .vscode/settings.json << 'EOF'
{
  "files.associations": {
    "*.sky": "python"
  },
  "editor.tabSize": 2,
  "editor.insertSpaces": true,
  "editor.formatOnSave": false,
  "[sky]": {
    "editor.defaultFormatter": "sky.formatter"
  }
}
EOF

echo "✅ Created .vscode/settings.json"

# Create code snippets
cat > .vscode/sky.code-snippets << 'EOF'
{
  "SKY Function": {
    "prefix": "fn",
    "body": [
      "function ${1:name}($2)",
      "  $0",
      "end"
    ],
    "description": "SKY function definition"
  },
  "SKY Async Function": {
    "prefix": "afn",
    "body": [
      "async function ${1:name}($2)",
      "  $0",
      "end"
    ],
    "description": "SKY async function"
  },
  "SKY Class": {
    "prefix": "class",
    "body": [
      "class ${1:Name}",
      "  function init($2)",
      "    self.$3 = $3",
      "  end",
      "  ",
      "  function ${4:method}($5)",
      "    $0",
      "  end",
      "end"
    ],
    "description": "SKY class definition"
  },
  "SKY If": {
    "prefix": "if",
    "body": [
      "if ${1:condition}",
      "  $0",
      "end"
    ],
    "description": "SKY if statement"
  },
  "SKY For": {
    "prefix": "for",
    "body": [
      "for ${1:item} in ${2:collection}",
      "  $0",
      "end"
    ],
    "description": "SKY for loop"
  },
  "SKY Enum": {
    "prefix": "enum",
    "body": [
      "enum ${1:Name}",
      "  ${2:Variant1}",
      "  ${3:Variant2}",
      "end"
    ],
    "description": "SKY enum definition"
  },
  "SKY Match": {
    "prefix": "match",
    "body": [
      "match ${1:value}",
      "  ${2:Pattern1} => $3",
      "  ${4:Pattern2} => $5",
      "end"
    ],
    "description": "SKY match expression"
  }
}
EOF

echo "✅ Created .vscode/sky.code-snippets"
echo ""
echo "🎉 VS Code setup complete!"
echo ""
echo "Next steps:"
echo "  1. Restart VS Code"
echo "  2. Open any .sky file"
echo "  3. Enjoy syntax highlighting!"
echo "  4. Try snippets: type 'fn' and press Tab"
echo ""

