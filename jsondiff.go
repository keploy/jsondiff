package colorisediff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/tidwall/gjson"
)

// Range represents a range with a start and end index.
type colorRange struct {
	Start int // Start is the starting index of the range.
	End   int // End is the ending index of the range.
}

// Diff holds the colorized differences between the expected and actual JSON responses.
// Expected: The colorized string representing the differences in the expected JSON response.
// Actual: The colorized string representing the differences in the actual JSON response.
type Diff struct {
	Expected string
	Actual   string
}

// CompareJSON compares two JSON objects and returns the differences as colorized strings.
// json1: The first JSON object to compare.
// json2: The second JSON object to compare.
// noise: A map containing fields to ignore during the comparison.
// Returns a ColorizedResponse containing the colorized differences for the expected and actual JSON responses.
func CompareJSON(expectedJSON []byte, actualJSON []byte, noise map[string][]string, disableColor bool) (Diff, error) {
	color.NoColor = disableColor
	// Calculate the differences between the two JSON objects.
	diffString, err := calculateJSONDiffs(expectedJSON, actualJSON)
	if err != nil || diffString == "" {
		return Diff{}, err
	}
	// Extract the modified keys from the diff string.
	modifiedKeys := extractKey(diffString)

	// Check if the modified keys exist in the provided maps and add additional context if they do.
	contextInfo, exists, error := checkKeyInMaps(expectedJSON, actualJSON, modifiedKeys)

	if error != nil {
		return Diff{}, error
	}

	if exists {
		diffString = contextInfo + "\n" + diffString
	}
	// Separate and colorize the diff string into expected and actual outputs.
	expect, actual := separateAndColorize(diffString, noise)

	return Diff{
		Expected: expect,
		Actual:   actual,
	}, nil
}

// Compare takes expected and actual JSON strings and returns the colorized differences.
// expectedJSON: The JSON string containing the expected values.
// actualJSON: The JSON string containing the actual values.
// Returns a Diff struct containing the colorized differences for the expected and actual JSON responses.
func Compare(expectedJSON, actualJSON string) Diff {
	// Calculate the ranges for differences between the expected and actual JSON strings.
	offsetExpected, offsetActual, _ := diffArrayRange(expectedJSON, actualJSON)

	// Define colors for highlighting differences.
	highlightExpected := color.FgHiRed
	highlightActual := color.FgHiGreen

	// Colorize the differences in the expected and actual JSON strings.
	colorizedExpected := breakSliceWithColor(expectedJSON, &highlightExpected, offsetExpected)
	colorizedActual := breakSliceWithColor(actualJSON, &highlightActual, offsetActual)

	// Return the colorized differences in a Diff struct.
	return Diff{
		Expected: colorizedExpected,
		Actual:   colorizedActual,
	}
}

// checkKeyInMaps checks if the given key exists in both JSON maps and returns additional context if found.
// expectedJSONMap: The first JSON map in byte form.
// actualJSONMap: The second JSON map in byte form.
// key: The key to check for existence in both maps.
// Returns a string with additional context and a boolean indicating if the key was found in both maps.
func checkKeyInMaps(expectedJSONMap, actualJSONMap []byte, targetKey string) (string, bool, error) {
	var expectedMap, actualMap map[string]interface{}

	// Unmarshal both JSON maps into Go maps.
	if err := json.Unmarshal(expectedJSONMap, &expectedMap); err != nil {
		fmt.Println("Error unmarshalling expected JSON map", string(expectedJSONMap), err)
		return "", false, err
	}
	if err := json.Unmarshal(actualJSONMap, &actualMap); err != nil {
		fmt.Println("Error unmarshalling actual JSON map")
		return "", false, err
	}

	// Iterate over the key-value pairs in the expected map.
	for key, expectedValue := range expectedMap {
		// Check if the key exists in the actual map, is not part of the provided key string, and values are deeply equal.
		if actualValue, exists := actualMap[key]; exists && !strings.Contains(targetKey, key) && reflect.DeepEqual(expectedValue, actualValue) {
			return fmt.Sprintf("%v:%v", key, expectedValue), true, nil
		}
	}

	// If no matching key-value pair is found, return an empty string and false.
	return "", false, nil

}

// calculateJSONDiffs calculates the differences between two JSON objects and returns a diff string.
// expectedJSON: The first JSON object in byte form.
// actualJSON: The second JSON object in byte form.
// Returns a string representing the differences and an error if any.
func calculateJSONDiffs(expectedJSON, actualJSON []byte) (string, error) {
	// Parse both JSON objects.
	expectedResult := gjson.ParseBytes(expectedJSON)
	actualResult := gjson.ParseBytes(actualJSON)

	var diffs []string

	// Iterate over key-value pairs in the expected JSON and compare with the actual JSON.
	expectedResult.ForEach(func(key, expectedValue gjson.Result) bool {
		actualValue := actualResult.Get(key.String())
		if !actualValue.Exists() || expectedValue.String() != actualValue.String() {
			diffs = append(diffs, fmt.Sprintf("- \"%s\": %v", key, expectedValue))
			if actualValue.Exists() {
				diffs = append(diffs, fmt.Sprintf("+ \"%s\": %v", key, actualValue))
			}
		}
		return true
	})

	// Iterate over the key-value pairs in the actual JSON and add any missing keys from the expected JSON.
	actualResult.ForEach(func(key, actualValue gjson.Result) bool {
		if !expectedResult.Get(key.String()).Exists() {
			diffs = append(diffs, fmt.Sprintf("+ \"%s\": %v", key, actualValue))
		}
		return true
	})

	// Join the diffs into a single string separated by newlines.
	return strings.Join(diffs, "\n"), nil
}

// extractKey extracts the keys from the diff string.
// diffString: The input string representing the differences.
// Returns a string containing all the keys separated by a pipe character.
func extractKey(diffString string) string {
	diffLines := strings.Split(diffString, "\n") // Split the diff string into lines.
	var keys []string

	// Iterate over each line in the diff string.
	for _, line := range diffLines {
		// Remove the leading '-' or '+' and any surrounding spaces
		line = strings.TrimSpace(line[1:])

		if colonIndex := strings.Index(line, ":"); colonIndex != -1 {
			// Extract and clean up the key
			key := strings.Trim(line[:colonIndex], `"'`)
			keys = append(keys, key)
		}
		// Add the key to the list of keys.
	}

	// Join the keys into a single string separated by a pipe character.
	return strings.Join(keys, "|")
}

// writeKeyValuePair writes a key-value pair to a string builder with optional colorization.
// builder: The string builder to write the key-value pair to.
// key: The key to be written.
// value: The value to be written.
// indent: The indentation string to use for formatting.
// colorFunc: The function to apply color to the value, if provided.
func writeKeyValuePair(builder *strings.Builder, key string, value interface{}, indent string, applyColor func(a ...interface{}) string) {
	// Serialize the value to a pretty-printed JSON string.
	serializedValue, _ := json.MarshalIndent(value, "", "  ")
	formattedValue := string(serializedValue)

	// Check if a color function is provided and the value is not empty.
	if applyColor != nil && value != "" {
		formattedValue = applyColor(formattedValue)
	}

	// Write the key-value pair to the builder with or without colorization.
	builder.WriteString(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, formattedValue))

}

// compareAndColorizeSlices compares two slices and returns the differences as colorized strings.
// a: The first slice to compare.
// b: The second slice to compare.
// indent: The indentation string to use for formatting.
// red, green: Functions to apply red and green colors respectively for differences.
// Returns two strings: the colorized differences for the expected and actual slices.
func compareAndColorizeSlices(a, b []interface{}, indent string, red, green func(a ...interface{}) string) (string, string) {
	var expectedOutput strings.Builder // Builder for the expected output string.
	var actualOutput strings.Builder   // Builder for the actual output string.
	maxLength := len(a)                // Determine the maximum length between the two slices.
	if len(b) > maxLength {
		maxLength = len(b)
	}

	// Iterate over the elements of the slices up to the maximum length.
	for i := 0; i < maxLength; i++ {
		var aValue, bValue interface{}
		aExists, bExists := i < len(a), i < len(b) // Flags to indicate if values exist in both slices

		// Assign the current element from the first slice if within bounds.
		if aExists {
			aValue = a[i]
		}

		// Assign the current element from the second slice if within bounds.
		if bExists {
			bValue = b[i]
		}

		// Use a switch to handle the cases based on the existence of values in both slices.
		switch {
		case !aExists && !bExists:
			// If neither value exists, continue the loop.
			continue

		case !aExists:
			// Only the second slice has a value.
			actualOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, green(serialize(bValue))))

		case !bExists:
			// Only the first slice has a value.
			expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, red(serialize(aValue))))

		default:
			// If both elements exist, compare and colorize them.
			switch v1 := aValue.(type) {
			case map[string]interface{}:
				if v2, ok := bValue.(map[string]interface{}); ok {
					// Recursively compare and colorize maps.
					expectedText, actualText := compareAndColorizeMaps(v1, v2, indent+"  ", red, green)
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, expectedText))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, actualText))
					continue
				}

			case []interface{}:
				if v2, ok := bValue.([]interface{}); ok {
					// Recursively compare and colorize slices.
					expectedText, actualText := compareAndColorizeSlices(v1, v2, indent+"  ", red, green)
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: [\n%s%s]\n", indent, i, expectedText, indent))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: [\n%s%s]\n", indent, i, actualText, indent))
					continue
				}

			default:
				// If values are not deeply equal, write the values with colors.
				if reflect.DeepEqual(aValue, bValue) {
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %v\n", indent, i, aValue))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: %v\n", indent, i, bValue))
					continue
				}
			}
			// If the values are not equal, colorize them.
			expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, red(serialize(aValue))))
			actualOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, green(serialize(bValue))))
		}
	}

	// Return the resulting colorized differences for the expected and actual slices.
	return expectedOutput.String(), actualOutput.String()
}

// serialize serializes a value to a pretty-printed JSON string.
func serialize(value interface{}) string {
	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "error"
	}
	return string(bytes)
}

// compare compares two values and writes the differences to the provided builders with optional colorization.
// key: The key associated with the values being compared.
// val1: The first value to compare.
// val2: The second value to compare.
// indent: The indentation string to use for formatting.
// expect: The builder for the expected output.
// actual: The builder for the actual output.
// red, green: Functions to apply red and green colors respectively for differences.
func compare(key string, val1, val2 interface{}, indent string, expect, actual *strings.Builder, red, green func(a ...interface{}) string) {
	switch v1 := val1.(type) {
	// Case for map[string]interface{} type
	case map[string]interface{}:
		// Check if the second value is also a map[string]interface{}
		if v2, ok := val2.(map[string]interface{}); ok {
			// Recursively compare and colorize maps
			expectedText, actualText := compareAndColorizeMaps(v1, v2, indent+"  ", red, green)
			expect.WriteString(fmt.Sprintf("%s\"%s\": %s\n", indent, key, expectedText))
			actual.WriteString(fmt.Sprintf("%s\"%s\": %s\n", indent, key, actualText))
			return
		}
		// If types do not match, write the key-value pairs with colors
		writeKeyValuePair(expect, key, val1, indent, red)
		writeKeyValuePair(actual, key, val2, indent, green)

	// Case for []interface{} type
	case []interface{}:
		// Check if the second value is also a []interface{}
		if v2, ok := val2.([]interface{}); ok {
			// Recursively compare and colorize slices
			expectedText, actualText := compareAndColorizeSlices(v1, v2, indent+"  ", red, green)
			expect.WriteString(fmt.Sprintf("%s\"%s\": [\n%s\n%s]\n", indent, key, expectedText, indent))
			actual.WriteString(fmt.Sprintf("%s\"%s\": [\n%s\n%s]\n", indent, key, actualText, indent))
			return
		}
		// If types do not match, write the key-value pairs with colors
		writeKeyValuePair(expect, key, val1, indent, red)
		writeKeyValuePair(actual, key, val2, indent, green)

	// Default case for other types
	default:
		// Check if the values are not deeply equal
		if !reflect.DeepEqual(val1, val2) {
			// Marshal values to pretty-printed JSON strings
			val1Str, err := json.MarshalIndent(val1, "", "  ")
			if err != nil {
				return
			}
			val2Str, err := json.MarshalIndent(val2, "", "  ")
			if err != nil {
				return
			}
			// Colorize the differences in the values
			c := color.FgRed
			offsetsStr1, offsetsStr2, _ := diffArrayRange(string(val1Str), string(val2Str))
			expectDiff := breakSliceWithColor(string(val1Str), &c, offsetsStr1)
			c = color.FgGreen
			actualDiff := breakSliceWithColor(string(val2Str), &c, offsetsStr2)
			expect.WriteString(breakLines(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, string(expectDiff))))
			actual.WriteString(breakLines(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, string(actualDiff))))
			return
		}
		// If values are equal, write the value without color
		valStr, err := json.MarshalIndent(val1, "", "  ")
		if err != nil {
			return
		}
		expect.WriteString(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, string(valStr)))
		actual.WriteString(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, string(valStr)))

	}
}

// separateAndColorize separates the diff string into expected and actual strings, applying color where appropriate.
// diffStr: The input string representing the differences.
// noise: A map containing noise elements to be ignored during processing.
// Returns two strings: the colorized expected and actual differences.
func separateAndColorize(diffStr string, noise map[string][]string) (string, string) {
	lines := strings.Split(diffStr, "\n") // Split the diff string into lines.
	lines = insertEmptyLines(lines)       // Insert empty lines between consecutive elements with the same symbol.
	// Initialize maps and arrays to store the expected and actual values.
	expectMap := make(map[string]interface{}, 0)
	actualMap := make(map[string]interface{}, 0)
	expectsArray := make([]interface{}, 0)
	actualsArray := make([]interface{}, 0)
	var isExpectMap, isActualMap bool
	expect, actual := "", ""

	// Iterate over the lines, processing each line and the next line together.
	for i := 0; i < len(lines)-1; i++ {
		var expectKey, actualKey string
		line := lines[i]
		nextLine := lines[i+1]

		// Process lines that start with a '-' indicating expected differences.
		if len(line) > 0 && line[0] == '-' && i != len(lines)-1 {
			if len(nextLine) > 3 && len(strings.SplitN(nextLine[3:], ":", 2)) == 2 {
				actualTrimmedLine := nextLine[3:] // Trim the '+ ' prefix from the next line.
				actualKeyValue := strings.SplitN(actualTrimmedLine, ":", 2)
				actualKey = strings.TrimSpace(actualKeyValue[0])
				value := strings.TrimSpace(actualKeyValue[1])
				var jsonObj map[string]interface{}
				switch {
				case json.Unmarshal([]byte(value), &jsonObj) == nil:
					isActualMap = true
					actualMap = map[string]interface{}{actualKey[:len(actualKey)-1]: jsonObj}
				case json.Unmarshal([]byte(value), &actualsArray) == nil:
				}

			}
			if len(strings.SplitN(line[3:], ":", 2)) == 2 {
				expectTrimmedLine := line[3:] // Trim the '- ' prefix from the current line.
				expectkeyValue := strings.SplitN(expectTrimmedLine, ":", 2)
				expectKey = strings.TrimSpace(expectkeyValue[0])
				value := strings.TrimSpace(expectkeyValue[1])
				var jsonObj map[string]interface{}
				switch {
				case json.Unmarshal([]byte(value), &jsonObj) == nil:
					isExpectMap = true
					expectMap = map[string]interface{}{expectKey[:len(expectKey)-1]: jsonObj}
				case json.Unmarshal([]byte(value), &expectsArray) == nil:
				}
			}
			// Define color functions for red and green.
			red := color.New(color.FgRed).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			var expectedText, actualText string
			// Compare and colorize maps or arrays.
			if !isExpectMap || !isActualMap {
				if actualKey != expectKey {
					continue
				}
				expectedText, actualText = compareAndColorizeSlices(expectsArray, actualsArray, " ", red, green)
			}

			if isExpectMap && isActualMap {
				expectedText, actualText = compareAndColorizeMaps(expectMap, actualMap, " ", red, green)
			}

			// Truncate and break lines to match with ellipsis.
			expectOutput, actualOutput := truncateToMatchWithEllipsis(breakLines(expectedText), breakLines(actualText))
			expect += breakLines(expectOutput) + "\n"
			actual += breakLines(actualOutput) + "\n"
			// Reset maps for the next iteration.
			expectMap = make(map[string]interface{}, 0)
			actualMap = make(map[string]interface{}, 0)

			// Remove processed lines from diffStr.
			diffStr = strings.Replace(diffStr, line, "", 1)
			diffStr = strings.Replace(diffStr, nextLine, "", 1)
		}
	}

	// If diffStr is empty, return the accumulated expected and actual strings.
	if diffStr == "" {
		return expect, actual
	}

	// Process remaining lines in diffStr.
	diffLines := strings.Split(diffStr, "\n")
	for i, line := range diffLines {
		if len(line) == 0 {
			continue
		}
		noised := false

		// Check for noise elements and adjust lines accordingly.
		for e := range noise {
			if strings.Contains(line, e) {
				if line[0] == '-' {
					line = " " + line[1:]
					expect += breakWithColor(line, nil, []colorRange{})
				} else if line[0] == '+' {
					line = " " + line[1:]
					actual += breakWithColor(line, nil, []colorRange{})
				}
				noised = true
				break
			}
		}

		if noised {
			continue
		}

		// Process lines that start with '-' indicating expected differences.
		// Determine if line starts with '-' or '+'
		switch line[0] {
		case '-':
			c := color.FgRed
			if i < len(diffLines)-1 && len(line) > 1 && diffLines[i+1] != "" && diffLines[i+1][0] == '+' {
				offsets, _ := diffIndexRange(line[1:], diffLines[i+1][1:])
				expect += breakWithColor(line, &c, offsets)
				continue
			}
			expect += breakWithColor(line, &c, []colorRange{{Start: 0, End: len(line)}})

		case '+':
			c := color.FgGreen
			if i > 0 && len(line) > 1 && diffLines[i-1] != "" && diffLines[i-1][0] == '-' {
				offsets, _ := diffIndexRange(line[1:], diffLines[i-1][1:])
				actual += breakWithColor(line, &c, offsets)
				continue
			}
			actual += breakWithColor(line, &c, []colorRange{{Start: 0, End: len(line)}})

		default:
			// Process lines that do not start with '-' or '+'
			expect += breakWithColor(line, nil, []colorRange{})
			actual += breakWithColor(line, nil, []colorRange{})
		}

	}

	// Return the accumulated expected and actual strings.
	return expect, actual
}

// breakWithColor applies color to specific ranges within the input string and breaks the string into lines.
// input: The string to be processed.
// c: The color attribute to apply to the specified ranges. If nil, no color is applied.
// highlightRanges: A slice of Range structs specifying the start and end indices for color application.
func breakWithColor(input string, c *color.Attribute, highlightRanges []colorRange) string {
	// Default paint function does nothing.
	paint := func(_ ...interface{}) string { return "" }
	// If a color attribute is provided, update the paint function to apply that color.
	if c != nil {
		paint = color.New(*c).SprintFunc()
	}
	var output strings.Builder // Use strings.Builder for efficient string concatenation.
	var isColorRange bool
	lineLen := 0

	// Iterate over each character in the input string.
	for i, char := range input {
		isColorRange = false
		// Check if the current index falls within any of the highlight ranges.
		for _, r := range highlightRanges {
			// Adjusted the range to be inclusive.
			if i >= r.Start && i < r.End {
				isColorRange = true
				break
			}
		}

		// Apply color if within a highlight range, otherwise add the character as is.
		if isColorRange {
			output.WriteString(paint(string(char)))
		} else {
			output.WriteString(string(char))
		}

		lineLen++
		// Break the line if it reaches the maximum line length.
		if lineLen == maxLineLength {
			output.WriteString("\n")
			lineLen = 0
		}
	}

	// Ensure the final output ends with a newline if there are remaining characters.
	if lineLen > 0 {
		output.WriteString("\n")
	}

	return output.String()
}

// isControlCharacter checks if a character is a control character.
func isControlCharacter(char rune) bool {
	return char < ' '
}

// maxLineLength is the maximum length of a line before it is wrapped.
const maxLineLength = 50

// breakLines breaks the input string into lines of a specified maximum length.
// input: The string to be processed and broken into lines.
// Returns the input string with line breaks inserted at the specified maximum length.
func breakLines(input string) string {
	var output strings.Builder      // Builder for the resulting output string.
	var currentLine strings.Builder // Builder for the current line being processed.
	lineLength := 0                 // Counter for the current line length.
	inANSISequence := false         // Boolean to track if we are inside an ANSI escape sequence.

	var ansiSequenceBuilder strings.Builder // Builder for the ANSI escape sequence.

	// Iterate over each character in the input string.
	for _, char := range input {
		switch {
		case inANSISequence: // We are currently inside an ANSI sequence
			ansiSequenceBuilder.WriteRune(char) // Add the character to the ANSI sequence builder
			if char == 'm' {                    // Check if the ANSI escape sequence has ended
				inANSISequence = false                                // Reset the flag
				currentLine.WriteString(ansiSequenceBuilder.String()) // Add the completed ANSI sequence to the current line
				ansiSequenceBuilder.Reset()                           // Reset the ANSI sequence builder
			}
		case char == '\x1b': // Start of an ANSI sequence
			inANSISequence = true
			ansiSequenceBuilder.WriteRune(char) // Add the start of the ANSI sequence to the builder
		case isControlCharacter(char) && char != '\n':
			currentLine.WriteRune(char) // Add control characters directly to the current line
		case lineLength >= maxLineLength:
			output.WriteString(currentLine.String()) // Add the current line to the output
			output.WriteRune('\n')                   // Add a newline character
			currentLine.Reset()                      // Reset the current line builder
			lineLength = 0                           // Reset the line length counter
		case char == '\n':
			output.WriteString(currentLine.String()) // Add the current line to the output
			output.WriteRune(char)                   // Add the newline character
			currentLine.Reset()                      // Reset the current line builder
			lineLength = 0                           // Reset the line length counter
		default:
			currentLine.WriteRune(char) // Add the character to the current line
			lineLength++                // Increment the line length counter
		}
	}

	if currentLine.Len() > 0 {
		output.WriteString(currentLine.String()) // Add the remaining characters in the current line to the output.
	}
	return output.String() // Return the processed output string.
}

// insertEmptyLines inserts empty lines between consecutive elements with the same symbol.
// lines: The input slice of strings to be processed.
// Returns a new slice of strings with empty lines inserted between consecutive elements with the same symbol.
func insertEmptyLines(lines []string) []string {
	var result []string // Initialize a slice to store the resulting lines.

	// Iterate over each line in the input slice.
	for i := 0; i < len(lines); i++ {
		result = append(result, lines[i]) // Append the current line to the result slice.

		// Check if the current line and the next line start with the same symbol.
		if i < len(lines)-1 && lines[i] != "" && lines[i][0] == lines[i+1][0] {
			result = append(result, "") // Insert an empty line between consecutive elements with the same symbol.
		}
	}

	// Return the result slice with inserted empty lines.
	return result
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

var ansiResetCode = "\x1b[0m"

// truncateToMatchWithEllipsis truncates the input strings to a specified length, adding ellipses in the middle.
// expectedText: The input string representing the expected text.
// actualText: The input string representing the actual text.
// Returns two strings: the truncated versions of the expected and actual texts.
func truncateToMatchWithEllipsis(expectedText, actualText string) (string, string) {
	expectedLines := strings.Split(expectedText, "\n") // Split the expected text into lines.
	actualLines := strings.Split(actualText, "\n")     // Split the actual text into lines.

	// Calculate the average number of lines between the expected and actual texts.
	matchLineCount := (len(expectedLines) + len(actualLines)) / 2

	// Define ANSI color codes for yellow, green, reset, and red.
	const yellow = "\033[33m"
	const green = "\033[32m"
	const reset = "\033[0m"
	const red = "\033[31m"

	// Build the ellipsis string with yellow color.
	var builder strings.Builder
	builder.WriteString(yellow)
	builder.WriteString(".\n")
	builder.WriteString(".\n")
	builder.WriteString(".")
	builder.WriteString(reset)
	ellipsis := builder.String()

	// Function to truncate the lines and add ellipses in the middle.
	truncate := func(lines []string, matchLineCount int, color string) string {
		// If the number of lines is less than or equal to the match line count, return the lines as a single string.
		if len(lines) <= matchLineCount {
			return strings.Join(lines, "\n")
		}

		// If the match line count is too small or the remaining lines are too few, return the lines as a single string.
		if matchLineCount <= 3 || len(lines)-matchLineCount < 3 {
			return strings.Join(lines, "\n")
		}

		// Calculate the number of lines for the top and bottom halves.
		topHalfLineCount := (matchLineCount - 3) / 2
		bottomHalfLineCount := matchLineCount - 3 - topHalfLineCount

		// Truncate the lines by keeping the top and bottom halves and adding ellipses in the middle.
		truncated := append(lines[:topHalfLineCount], ellipsis+color)
		truncated = append(truncated, lines[len(lines)-bottomHalfLineCount:]...)
		return strings.Join(truncated, "\n") + reset
	}

	// Truncate the expected and actual lines using the truncate function.
	truncatedExpected := truncate(expectedLines, matchLineCount+1, red)
	truncatedActual := truncate(actualLines, matchLineCount+1, green)

	// Return the truncated versions of the expected and actual texts.
	return truncatedExpected, truncatedActual
}

// compareAndColorizeMaps compares two maps and returns the differences as colorized strings.
// a: The first map to compare.
// b: The second map to compare.
// indent: The indentation string to use for formatting.
// red, green: Functions to apply red and green colors respectively.
// Returns two strings: the colorized differences for the expected and actual maps.
func compareAndColorizeMaps(a, b map[string]interface{}, indent string, red, green func(a ...interface{}) string) (string, string) {
	var expectedOutput, actualOutput strings.Builder // Builders for the resulting strings.
	expectedOutput.WriteString("{\n")                // Start the expected output with an opening brace and newline.
	actualOutput.WriteString("{\n")                  // Start the actual output with an opening brace and newline.

	// Iterate over each key-value pair in the first map.
	for key, aValue := range a {
		bValue, bHasKey := b[key] // Get the corresponding value from the second map and check if the key exists.
		if !bHasKey {             // If the key does not exist in the second map.
			writeKeyValuePair(&expectedOutput, red(key), aValue, indent+"  ", red) // Write the key-value pair with red color.
			continue                                                               // Move to the next key-value pair.
		}

		// Compare the values for the current key in both maps.
		compare(key, aValue, bValue, indent+"  ", &expectedOutput, &actualOutput, red, green)
	}

	// Iterate over each key-value pair in the second map.
	for key, bValue := range b {
		if _, aHasKey := a[key]; !aHasKey { // If the key does not exist in the first map.
			writeKeyValuePair(&actualOutput, green(key), bValue, indent+"  ", green) // Write the key-value pair with green color.
		}
	}

	expectedOutput.WriteString(indent + "}") // Close the expected output with a closing brace.
	actualOutput.WriteString(indent + "}")   // Close the actual output with a closing brace.

	// Return the resulting strings for the expected and actual maps.
	return expectedOutput.String(), actualOutput.String()
}

// CompareHeaders compares the headers of the expected and actual maps and returns the differences as colorized strings.
// expect: The map containing the expected header values.
// actual: The map containing the actual header values.
// Returns a ColorizedResponse containing the colorized differences for the expected and actual headers.
func CompareHeaders(expectedHeaders, actualHeaders map[string]string) Diff {
	var expectAll, actualAll strings.Builder // Builders for the resulting strings.

	// Iterate over each key-value pair in the expected map.
	for key, expValue := range expectedHeaders {
		actValue := actualHeaders[key] // Get the corresponding value from the actual map.

		// Calculate the offsets of the differences between the expected and actual values.
		offsetsStr1, offsetsStr2, _ := diffArrayRange(string(expValue), string(actValue))

		// Define colors for highlighting differences.
		cE, cA := color.FgHiRed, color.FgHiGreen

		// Colorize the differences in the expected and actual values.
		expectDiff := key + ": " + breakSliceWithColor(string(expValue), &cE, offsetsStr1)
		actualDiff := key + ": " + breakSliceWithColor(string(actValue), &cA, offsetsStr2)

		// Add the colorized differences to the builders.
		expectAll.WriteString(breakLines(expectDiff) + "\n")
		actualAll.WriteString(breakLines(actualDiff) + "\n")
	}

	// Return the resulting strings.
	return Diff{Expected: expectAll.String(), Actual: actualAll.String()}
}

// breakSliceWithColor breaks the input string into slices and applies color to specified offsets.
// s: The input string to be processed.
// c: The color attribute to apply to the specified offsets.
// offsets: A slice of indices specifying which words to colorize.
func breakSliceWithColor(s string, c *color.Attribute, offsets []int) string {
	var result strings.Builder                  // Use strings.Builder for efficient string concatenation.
	coloredString := color.New(*c).SprintFunc() // Function to apply the specified color.
	words := strings.Split(s, " ")              // Split the input string into words.

	// Iterate over each word in the slice.
	for i, word := range words {
		// Check if the current index is in the offsets slice.
		if contains(offsets, i) {
			// If it is, apply the color to the word and append it to the result.
			result.WriteString(coloredString(word) + " ")
			continue
		}
		// If it isn't, append the word as-is to the result.
		result.WriteString(word + " ")
	}

	return result.String() // Return the concatenated result as a string.
}

// contains checks if a slice contains a specific element.
// It returns true if the element is found in the slice, otherwise false.
func contains(slice []int, element int) bool {
	// Iterate over each element in the slice.
	for _, e := range slice {
		// If the current element matches the target element, return true.
		if e == element {
			return true
		}
	}
	// If the loop completes without finding the element, return false.
	return false
}

// diffIndexRange calculates the ranges of differences between two strings of words.
// It returns a slice of colorRange structs indicating the start and end indices of differences and a boolean indicating if there are differences.
func diffIndexRange(str1, str2 string) ([]colorRange, bool) {
	var ranges []colorRange // Slice to hold the ranges of differences.
	hasDifference := false  // Boolean to track if there are any differences.

	// Split the input strings into slices of words.
	words1 := strings.Split(str1, " ")
	words2 := strings.Split(str2, " ")

	// Determine the maximum length between the two word slices.
	maxLen := len(words1)
	if len(words2) > maxLen {
		maxLen = len(words2)
	}

	startIndex := 0 // Initialize the starting index for the ranges.

	// Iterate over the words up to the maximum length.
	for i := 0; i < maxLen; i++ {
		var word1, word2 string

		switch {
		case i < len(words1) && i < len(words2):
			// Both strings have words at index i, compare them.
			word1 = words1[i]
			word2 = words2[i]
			if word1 != word2 {
				hasDifference = true
				// Calculate the end index for the differing word.
				endIndex := startIndex + len(word1)
				// Record the range of the differing words.
				ranges = append(ranges, colorRange{Start: startIndex, End: endIndex})
			}
		case i < len(words1):
			// Only the first string has a word at index i (i.e., words1 is longer).
			word1 = words1[i]
			hasDifference = true
			// Calculate the end index and record the range.
			endIndex := startIndex + len(word1)
			ranges = append(ranges, colorRange{Start: startIndex, End: endIndex})
		case i < len(words2):
			// Only the second string has a word at index i (i.e., words2 is longer).
			word2 = words2[i]
			hasDifference = true
			// This case does not add ranges from words2 because we are only recording ranges from words1.
		}

		// Update the starting index for the next word, accounting for space after each word.
		startIndex += len(word1) + 1
	}

	// Return the ranges of differences and the difference flag.
	return ranges, hasDifference
}

// diffArrayRange calculates the indices of differences between two strings of words.
// It returns the indices where the words differ in both strings, and a boolean indicating if there are differences.
func diffArrayRange(s1, s2 string) ([]int, []int, bool) {
	var indices1, indices2 []int // Slices to hold the indices of differences for each string.
	diffFound := false           // Boolean to track if there are any differences.

	// Split the input strings into slices of words.
	words1 := strings.Split(s1, " ")
	words2 := strings.Split(s2, " ")

	// Determine the maximum length between the two word slices.
	maxLen := len(words1)
	if len(words2) > maxLen {
		maxLen = len(words2)
	}

	// Iterate over the words up to the maximum length.
	for i := 0; i < maxLen; i++ {
		switch {
		case i < len(words1) && i < len(words2): // Both strings have a word at index i
			if words1[i] != words2[i] { // If words are different, record the indices
				indices1 = append(indices1, i)
				indices2 = append(indices2, i)
				diffFound = true
			}
		case i < len(words1): // Only the first string has a word at index i
			indices1 = append(indices1, i)
			diffFound = true
		case i < len(words2): // Only the second string has a word at index i
			indices2 = append(indices2, i)
			diffFound = true
		}
	}

	// Return the indices of differences for both strings and whether differences were found.
	return indices1, indices2, diffFound
}
