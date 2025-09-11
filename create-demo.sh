#!/bin/bash

# create-demo.sh - Create a new demo project based on the counter demo template
# Usage: ./create-demo.sh <demo_name>

set -e

# Check if demo name is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <demo_name>"
    echo "Example: $0 my_awesome_demo"
    exit 1
fi

DEMO_NAME="$1"
DEMO_DIR="examples/$DEMO_NAME"

# Validate demo name (alphanumeric and underscores only)
if [[ ! "$DEMO_NAME" =~ ^[a-zA-Z0-9_]+$ ]]; then
    echo "Error: Demo name must contain only alphanumeric characters and underscores"
    exit 1
fi

# Check if demo directory already exists
if [ -d "$DEMO_DIR" ]; then
    echo "Error: Demo directory '$DEMO_DIR' already exists"
    exit 1
fi

echo "Creating demo: $DEMO_NAME"
echo "Directory: $DEMO_DIR"

# Create demo directory
mkdir -p "$DEMO_DIR"

# Create index.html
cat > "$DEMO_DIR/index.html" << EOF
<!doctype html>
<html>
<head>
  <meta charset="utf-8"/>
  <title>UiwGo $DEMO_NAME</title>
  <style>
    @import "tailwindcss";
  </style>
</head>
<body>
  <div id="app"></div>

  <script src="wasm_exec.js"></script>
  <script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject).then((result) => {
      go.run(result.instance);
    });
  </script>
</body>
</html>
EOF

# Capitalize first letter of demo name for function names
DEMO_NAME_CAPITALIZED="$(echo ${DEMO_NAME:0:1} | tr '[:lower:]' '[:upper:]')${DEMO_NAME:1}"

# Create main.go
cat > "$DEMO_DIR/main.go" << EOF
//go:build js && wasm

package main

import (
	"fmt"

	"github.com/ozanturksever/logutil"
	comps "github.com/ozanturksever/uiwgo/comps"
	dom "github.com/ozanturksever/uiwgo/dom"
	reactivity "github.com/ozanturksever/uiwgo/reactivity"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	// Mount the app and get a disposer function
	// In a real app, you might want to store this disposer to clean up when needed
	disposer := comps.Mount("app", func() Node { return ${DEMO_NAME_CAPITALIZED}App() })
	_ = disposer // We don't use it in this example since the app runs indefinitely

	// Prevent exit
	select {}
}

func ${DEMO_NAME_CAPITALIZED}App() Node {
	// Create a simple signal for demonstration
	message := reactivity.CreateSignal("Hello from $DEMO_NAME!")

	// Effect logging to console
	reactivity.CreateEffect(func() {
		logutil.Log("Message changed:", message.Get())
	})

	// Setup DOM event handlers after mount
	comps.OnMount(func() {
		// Get DOM elements and bind events using the new DOM API
		if changeBtn := dom.GetElementByID("change-btn"); changeBtn != nil {
			dom.BindClickToCallback(changeBtn, func() {
				message.Set(fmt.Sprintf("Updated at %d", reactivity.CreateSignal(0).Get()))
			})
		}
	})

	return Div(
		Style("font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; background-color: #f5f5f5; min-height: 100vh;"),
		Div(
			Style("background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center;"),
			H1(Text("$DEMO_NAME Demo")),
			P(Text("A demo project created with create-demo.sh")),

			Div(
				ID("message-display"),
				Style("font-size: 1.5em; font-weight: bold; color: #333; margin: 20px 0; padding: 20px; background-color: #f8f9fa; border-radius: 8px; border: 2px solid #e9ecef;"),
				comps.BindText(func() string { return message.Get() }),
			),

			Div(
				Button(
					ID("change-btn"),
					Style("font-size: 1.2em; padding: 10px 20px; margin: 10px; border: none; border-radius: 5px; cursor: pointer; background-color: #007bff; color: white; transition: background-color 0.2s;"),
					Text("Change Message"),
				),
			),

			Div(
				Style("margin-top: 20px; color: #666; font-style: italic;"),
				Text("Click the button to update the message!"),
			),
		),
	)
}
EOF

# Create main_test.go
cat > "$DEMO_DIR/main_test.go" << 'EOF'
//go:build !js && !wasm

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func Test${DEMO_NAME_CAPITALIZED}App(t *testing.T) {
	// Create and start the development server
	server := testhelpers.NewViteServer("${DEMO_NAME}", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with default configuration
	config := testhelpers.DefaultConfig()
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	// Navigate to the test server and perform the test
	var messageText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the demo app
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and WASM to initialize
		chromedp.WaitVisible(`#message-display`, chromedp.ByID),
		chromedp.WaitVisible(`#change-btn`, chromedp.ByID),

		// Wait a bit more for WASM to fully initialize
		chromedp.Sleep(2*time.Second),

		// Get initial message
		chromedp.Text(`#message-display`, &messageText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Assert that the initial message is correct
	expected := "Hello from ${DEMO_NAME}!"
	if !strings.Contains(messageText, expected) {
		t.Errorf("Expected message text to contain '%s', but got: '%s'", expected, messageText)
	}

	t.Logf("Test passed! Initial message: %s", messageText)
}

func Test${DEMO_NAME_CAPITALIZED}ButtonClick(t *testing.T) {
	// Create and start the development server
	server := testhelpers.NewViteServer("${DEMO_NAME}", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with default configuration
	config := testhelpers.DefaultConfig()
	chromedpCtx := testhelpers.MustNewChromedpContext(config)
	defer chromedpCtx.Cancel()

	// Navigate to the test server and perform the test
	var initialText, updatedText string
	err := chromedp.Run(chromedpCtx.Ctx,
		// Navigate to the demo app
		chromedp.Navigate(server.URL()),

		// Wait for the page to load and WASM to initialize
		chromedp.WaitVisible(`#message-display`, chromedp.ByID),
		chromedp.Sleep(2*time.Second),

		// Get initial text
		chromedp.Text(`#message-display`, &initialText, chromedp.ByID),

		// Click the change button
		chromedp.Click(`#change-btn`, chromedp.ByID),
		chromedp.Sleep(500*time.Millisecond),

		// Get updated text
		chromedp.Text(`#message-display`, &updatedText, chromedp.ByID),
	)

	if err != nil {
		t.Fatalf("Browser automation failed: %v", err)
	}

	// Assert that the text changed
	if initialText == updatedText {
		t.Errorf("Expected text to change after button click, but it remained: '%s'", initialText)
	}

	t.Logf("Button click test passed! Initial: %s, Updated: %s", initialText, updatedText)
}
EOF

# Replace placeholders in main_test.go using a temporary file approach
temp_file=$(mktemp)
cat "$DEMO_DIR/main_test.go" | sed "s/\${DEMO_NAME_CAPITALIZED}/$DEMO_NAME_CAPITALIZED/g" | sed "s/\${DEMO_NAME}/$DEMO_NAME/g" > "$temp_file"
mv "$temp_file" "$DEMO_DIR/main_test.go"

# Create vite.config.js
cat > "$DEMO_DIR/vite.config.js" << EOF
import viteCfgFactory, { parseCliArgs } from "../../vite.config.js";
import { resolve } from "path";

const { prod } = parseCliArgs();

export default viteCfgFactory(
    resolve(import.meta.dirname, "index.html"),
    "dist",
    [{
        input: resolve(import.meta.dirname, ".") + "/**",
        output: "/"
    }],
    [\`GOOS=js GOARCH=wasm go build \${prod ? '-ldflags="-s -w"' : ''} -o examples/$DEMO_NAME/main.wasm examples/$DEMO_NAME/main.go\`],
    "$DEMO_NAME"
);
EOF

echo "Created demo files in $DEMO_DIR"

# Update Makefile - add dev target
echo "Updating Makefile..."
if ! grep -q "dev-$DEMO_NAME:" Makefile; then
    # Add the new dev target to the end of the file
    printf "\n\ndev-$DEMO_NAME:\n\t@echo \"==> Starting Vite dev server for $DEMO_NAME example...\"\n\tnpm run dev:$DEMO_NAME\n" >> Makefile
fi

# Update package.json - add dev script
echo "Updating package.json..."
if ! grep -q "dev:$DEMO_NAME" package.json; then
    # Use a simple approach: find the closing brace of scripts and insert before it
    # First backup the original
    cp package.json package.json.bak
    
    # Use awk to insert the new script before the closing brace of scripts section
    awk -v demo="$DEMO_NAME" '
    BEGIN { in_scripts = 0; last_dev_line = 0 }
    /"scripts": {/ { in_scripts = 1 }
    in_scripts && /^    "dev:.*"/ { last_dev_line = NR }
    /^  },/ && in_scripts {
        print "    \"dev:" demo "\": \"vite -c examples/" demo "/vite.config.js\""
        in_scripts = 0
    }
    {
        if (NR == last_dev_line && in_scripts && !/,$/) {
            gsub(/"$/, ",")
        }
        print
    }
    ' package.json.bak > package.json
    
    # Clean up backup
    rm package.json.bak
fi

echo ""
echo "✅ Demo '$DEMO_NAME' created successfully!"
echo ""
echo "📁 Files created:"
echo "   - $DEMO_DIR/index.html"
echo "   - $DEMO_DIR/main.go"
echo "   - $DEMO_DIR/main_test.go"
echo "   - $DEMO_DIR/vite.config.js"
echo ""
echo "📝 Updated:"
echo "   - Makefile (added dev-$DEMO_NAME target)"
echo "   - package.json (added dev:$DEMO_NAME script)"
echo ""
echo "🚀 To run your demo:"
echo "   timeout 5s make run $DEMO_NAME"
echo "   # or"
echo "   timeout 5s make dev-$DEMO_NAME"
echo ""
echo "🧪 To test your demo:"
echo "   make test-example $DEMO_NAME"
echo "   # or"
echo "   make test-$DEMO_NAME"
echo ""
echo "🔨 To build your demo:"
echo "   make build $DEMO_NAME"
echo ""
echo "Happy coding! 🎉"