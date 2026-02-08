package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)

func main() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("      DIRECT GOJA LIBRARY LOADING TEST")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Create goja runtime
	fmt.Println("[1] Creating goja runtime...")
	runtime := goja.New()
	fmt.Println("✓ Runtime created")

	// Load Lodash from cache
	fmt.Println()
	fmt.Println("[2] Loading Lodash library from cache...")
	lodashPath := filepath.Join(".vm-cache", "libraries", "lodash-4.17.21.js")
	
	content, err := os.ReadFile(lodashPath)
	if err != nil {
		fmt.Printf("✗ Failed to read Lodash: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Read Lodash (%d bytes)\n", len(content))

	// Execute Lodash in runtime
	fmt.Println()
	fmt.Println("[3] Executing Lodash code in goja runtime...")
	if _, err := runtime.RunString(string(content)); err != nil {
		fmt.Printf("✗ Failed to execute Lodash: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Lodash executed")

	// Test if Lodash is available
	fmt.Println()
	fmt.Println("[4] Testing if Lodash (_ ) is available...")
	testCode := `
		if (typeof _ === 'undefined') {
			throw new Error("Lodash is not defined!");
		}
		"Lodash is available!";
	`
	result, err := runtime.RunString(testCode)
	if err != nil {
		fmt.Printf("✗ Lodash check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %s\n", result.String())

	// Test Lodash functionality
	fmt.Println()
	fmt.Println("[5] Testing Lodash functions...")
	functionalityTest := `
		const numbers = [1, 2, 3, 4, 5];
		const doubled = _.map(numbers, n => n * 2);
		const chunked = _.chunk([1, 2, 3, 4, 5, 6], 2);
		const result = {
			doubled: doubled,
			chunked: chunked,
			version: _.VERSION
		};
		JSON.stringify(result);
	`
	result, err = runtime.RunString(functionalityTest)
	if err != nil {
		fmt.Printf("✗ Lodash functionality test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Lodash functions work: %s\n", result.String())

	// Summary
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("                 ALL TESTS PASSED! ✓")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("KEY FINDINGS:")
	fmt.Println("1. ✓ Goja runtime created successfully")
	fmt.Println("2. ✓ Lodash library loaded from cache")
	fmt.Println("3. ✓ Lodash executed in goja runtime")
	fmt.Println("4. ✓ Lodash global (_ ) is available")
	fmt.Println("5. ✓ Lodash functions (map, chunk) work correctly")
	fmt.Println()
	fmt.Println("CONCLUSION: Library loading into goja runtime is WORKING!")
	fmt.Println()
}
