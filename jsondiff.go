package jsonDiff

import (
	"bufio"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/tidwall/gjson"
)

// Range represents a range with a start and end index.
type Range struct {
	Start int // Start is the starting index of the range.
	End   int // End is the ending index of the range.
}

// SprintJSONDiff compares two JSON objects and returns the differences as colorized strings.
func SprintJSONDiff(json1 []byte, json2 []byte, noise map[string][]string) (error, string, string) {
	// Calculate the differences between the two JSON objects.
	diffString, err := calculateJSONDiffs(json1, json2)
	if err != nil || diffString == "" {
		return err, "", ""
	}

	// Extract the modified keys from the diff string.
	modifiedKeys := extractKey(diffString)

	// Check if the modified keys exist in the provided maps and add additional context if they do.
	additionalContext, exists := checkKeyInMaps(json1, json2, modifiedKeys)
	if exists {
		diffString = additionalContext + "\n" + diffString
	}

	// Separate and colorize the diff string into expected and actual outputs.
	expect, actual := separateAndColorize(diffString, noise)

	return nil, expect, actual
}

// SprintDiff takes expected and actual JSON strings and returns the colorized differences.
func SprintDiff(expect, actual string) (string, string) {
	// Calculate the ranges for differences between the expected and actual JSON strings.
	offsetExpect, offsetActual, _ := diffArrayRange(expect, actual)

	// Define colors for highlighting differences.
	cE, cA := color.FgHiRed, color.FgHiGreen

	var exp, act string

	// Colorize the differences in the expected and actual JSON strings.
	exp += breakSliceWithColor(expect, &cE, offsetExpect)
	act += breakSliceWithColor(actual, &cA, offsetActual)
	return exp, act
}

// checkKeyInMaps checks if the given key exists in both JSON maps and returns additional context if found.
// jsonMap1: The first JSON map in byte form.
// jsonMap2: The second JSON map in byte form.
// key: The key to check for existence in both maps.
// Returns a string with additional context and a boolean indicating if the key was found in both maps.
func checkKeyInMaps(jsonMap1, jsonMap2 []byte, key string) (string, bool) {
	var map1, map2 map[string]interface{}

	// Unmarshal the first JSON map into a Go map.
	if err := json.Unmarshal(jsonMap1, &map1); err != nil {
		return "", false // Return false if unmarshalling fails.
	}

	// Unmarshal the second JSON map into a Go map.
	if err := json.Unmarshal(jsonMap2, &map2); err != nil {
		return "", false // Return false if unmarshalling fails.
	}

	// Iterate over the key-value pairs in the first map.
	for k, v1 := range map1 {
		// Check if the key exists in the second map and is not in the provided key string.
		if v2, ok := map2[k]; ok && !strings.Contains(key, k) {
			// Check if the values are deeply equal.
			if reflect.DeepEqual(v1, v2) {
				return fmt.Sprintf("%v:%v", k, v1), true // Return the key-value pair and true if they are equal.
			}
		}
	}

	// Return an empty string and false if the key is not found in both maps or the values are not equal.
	return "", false
}

// calculateJSONDiffs calculates the differences between two JSON objects and returns a diff string.
// json1: The first JSON object in byte form.
// json2: The second JSON object in byte form.
// Returns a string representing the differences and an error if any.
func calculateJSONDiffs(json1 []byte, json2 []byte) (string, error) {
	result1 := gjson.ParseBytes(json1) // Parse the first JSON object.
	result2 := gjson.ParseBytes(json2) // Parse the second JSON object.

	var diffStrings []string

	// Iterate over the key-value pairs in the first JSON object.
	result1.ForEach(func(key, value gjson.Result) bool {
		value2 := result2.Get(key.String())
		// If the key does not exist in the second JSON object, add it to the diff.
		if !value2.Exists() {
			diffStrings = append(diffStrings, fmt.Sprintf("- \"%s\": %v", key, value))
		} else if value.String() != value2.String() { // If the values are different, add both to the diff.
			diffStrings = append(diffStrings, fmt.Sprintf("- \"%s\": %v", key, value))
			diffStrings = append(diffStrings, fmt.Sprintf("+ \"%s\": %v", key, value2))
		}
		return true
	})

	// Iterate over the key-value pairs in the second JSON object.
	result2.ForEach(func(key, value gjson.Result) bool {
		// If the key does not exist in the first JSON object, add it to the diff.
		if !result1.Get(key.String()).Exists() {
			diffStrings = append(diffStrings, fmt.Sprintf("+ \"%s\": %v", key, value))
		}
		return true
	})

	// Join the diff strings into a single string separated by newlines.
	return strings.Join(diffStrings, "\n"), nil
}

// extractKey extracts the keys from the diff string.
// diffString: The input string representing the differences.
// Returns a string containing all the keys separated by a pipe character.
func extractKey(diffString string) string {
	diffStrings := strings.Split(diffString, "\n") // Split the diff string into lines.
	var keys []string

	// Iterate over each line in the diff string.
	for _, str := range diffStrings {
		str = strings.TrimSpace(str[1:]) // Remove the leading '-' or '+' and trim spaces.

		colonIndex := strings.Index(str, ":") // Find the index of the colon.
		if colonIndex == -1 {
			continue // Skip if there is no colon.
		}

		key := strings.TrimSpace(str[:colonIndex]) // Extract the key up to the colon.
		key = strings.Trim(key, `"'`)              // Trim quotes from the key.
		keys = append(keys, key)                   // Add the key to the list of keys.
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
func writeKeyValuePair(builder *strings.Builder, key string, value interface{}, indent string, colorFunc func(a ...interface{}) string) {
	// Serialize the value to a pretty-printed JSON string.
	serializedValue, _ := json.MarshalIndent(value, "", "  ")
	valueStr := string(serializedValue)

	// Check if a color function is provided and the value is not empty.
	if colorFunc != nil && !reflect.DeepEqual(value, "") {
		// Write the key-value pair with colorization.
		builder.WriteString(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, colorFunc(valueStr)))
	} else {
		// Write the key-value pair without colorization.
		builder.WriteString(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, valueStr))
	}
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
		var aExists, bExists bool

		// Assign the current element from the first slice if within bounds.
		if i < len(a) {
			aValue = a[i]
			aExists = true
		}

		// Assign the current element from the second slice if within bounds.
		if i < len(b) {
			bValue = b[i]
			bExists = true
		}

		// If both elements exist, compare and colorize them.
		if aExists && bExists {
			switch v1 := aValue.(type) {
			case map[string]interface{}:
				if v2, ok := bValue.(map[string]interface{}); ok {
					// Recursively compare and colorize maps.
					expectedText, actualText := compareAndColorizeMaps(v1, v2, indent+"  ", red, green)
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, expectedText))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, actualText))
				} else {
					// If types do not match, write the values with colors.
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, red("%v", aValue)))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, green("%v", bValue)))
				}
			case []interface{}:
				if v2, ok := bValue.([]interface{}); ok {
					// Recursively compare and colorize slices.
					expectedText, actualText := compareAndColorizeSlices(v1, v2, indent+"  ", red, green)
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: [\n%s%s]\n", indent, i, expectedText, indent))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: [\n%s%s]\n", indent, i, actualText, indent))
				} else {
					// If types do not match, write the values with colors.
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, red("%v", aValue)))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, green("%v", bValue)))
				}
			default:
				if !reflect.DeepEqual(aValue, bValue) {
					// If values are not equal, write the values with colors.
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, red(aValue)))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, green(bValue)))
				} else {
					// If values are equal, write the values without color.
					expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %v\n", indent, i, aValue))
					actualOutput.WriteString(fmt.Sprintf("%s[%d]: %v\n", indent, i, bValue))
				}
			}
		} else if aExists {
			// If only the element from the first slice exists, write it with red color.
			expectedOutput.WriteString(fmt.Sprintf("%s[%d]: %s\n", indent, i, red(serialize(aValue))))
		} else if bExists {
			// If only the element from the second slice exists, write it with green color.
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
		} else {
			// If types do not match, write the key-value pairs with colors
			writeKeyValuePair(expect, key, val1, indent, red)
			writeKeyValuePair(actual, key, val2, indent, green)
		}
	// Case for []interface{} type
	case []interface{}:
		// Check if the second value is also a []interface{}
		if v2, ok := val2.([]interface{}); ok {
			// Recursively compare and colorize slices
			expectedText, actualText := compareAndColorizeSlices(v1, v2, indent+"  ", red, green)
			expect.WriteString(fmt.Sprintf("%s\"%s\": [\n%s\n%s]\n", indent, key, expectedText, indent))
			actual.WriteString(fmt.Sprintf("%s\"%s\": [\n%s\n%s]\n", indent, key, actualText, indent))
		} else {
			// If types do not match, write the key-value pairs with colors
			writeKeyValuePair(expect, key, val1, indent, red)
			writeKeyValuePair(actual, key, val2, indent, green)
		}
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
		} else {
			// If values are equal, write the value without color
			valStr, err := json.MarshalIndent(val1, "", "  ")
			if err != nil {
				return
			}
			expect.WriteString(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, string(valStr)))
			actual.WriteString(fmt.Sprintf("%s\"%s\": %s,\n", indent, key, string(valStr)))
		}
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
	expectsMap := make(map[string]interface{}, 0)
	actualsMap := make(map[string]interface{}, 0)
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
			if len(nextLine) > 3 {
				actualTrimmedLine := nextLine[3:] // Trim the '+ ' prefix from the next line.
				actualKeyValue := strings.SplitN(actualTrimmedLine, ":", 2)
				if len(actualKeyValue) == 2 {
					actualKey = strings.TrimSpace(actualKeyValue[0])
					value := strings.TrimSpace(actualKeyValue[1])

					var jsonObj map[string]interface{}
					err := json.Unmarshal([]byte(value), &jsonObj)
					if err != nil {
						var arrayObj []interface{}
						arrayError := json.Unmarshal([]byte(value), &arrayObj)
						if arrayError != nil {
							continue
						}
						actualsArray = arrayObj
					} else {
						isActualMap = true
						actualsMap = map[string]interface{}{actualKey[:len(actualKey)-1]: jsonObj}
					}
				}
			}

			expectTrimmedLine := line[3:] // Trim the '- ' prefix from the current line.
			expectkeyValue := strings.SplitN(expectTrimmedLine, ":", 2)
			if len(expectkeyValue) == 2 {
				expectKey = strings.TrimSpace(expectkeyValue[0])
				value := strings.TrimSpace(expectkeyValue[1])
				var jsonObj map[string]interface{}
				err := json.Unmarshal([]byte(value), &jsonObj)
				if err != nil {
					var arrayObj []interface{}
					arrayError := json.Unmarshal([]byte(value), &arrayObj)
					if arrayError != nil {
						continue
					}
					expectsArray = arrayObj
				} else {
					isExpectMap = true
					expectsMap = map[string]interface{}{expectKey[:len(expectKey)-1]: jsonObj}
				}
			}

			// Define color functions for red and green.
			red := color.New(color.FgRed).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			var expectedText, actualText string

			// Compare and colorize maps or arrays.
			if isExpectMap && isActualMap {
				expectedText, actualText = compareAndColorizeMaps(expectsMap, actualsMap, " ", red, green)
			} else if actualKey == expectKey {
				expectedText, actualText = compareAndColorizeSlices(expectsArray, actualsArray, " ", red, green)
			} else {
				continue
			}

			// Truncate and break lines to match with ellipsis.
			expectOutput, actualOutput := truncateToMatchWithEllipsis(breakLines(expectedText), breakLines(actualText))
			expect += breakLines(expectOutput) + "\n"
			actual += breakLines(actualOutput) + "\n"

			// Reset maps for the next iteration.
			expectsMap = make(map[string]interface{}, 0)
			actualsMap = make(map[string]interface{}, 0)

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
		if len(line) > 0 {
			noised := false

			// Check for noise elements and adjust lines accordingly.
			for e := range noise {
				if strings.Contains(line, e) {
					if line[0] == '-' {
						line = " " + line[1:]
						expect += breakWithColor(line, nil, []Range{})
					} else if line[0] == '+' {
						line = " " + line[1:]
						actual += breakWithColor(line, nil, []Range{})
					}
					noised = true
				}
			}

			if noised {
				continue
			}

			// Process lines that start with '-' indicating expected differences.
			if line[0] == '-' {
				c := color.FgRed
				if i < len(diffLines)-1 && len(line) > 1 && diffLines[i+1] != "" && diffLines[i+1][0] == '+' {
					offsets, _ := diffIndexRange(line[1:], diffLines[i+1][1:])
					expect += breakWithColor(line, &c, offsets)
				} else {
					expect += breakWithColor(line, &c, []Range{{Start: 0, End: len(line) - 1}})
				}
			} else if line[0] == '+' { // Process lines that start with '+' indicating actual differences.
				c := color.FgGreen
				if i > 0 && len(line) > 1 && diffLines[i-1] != "" && diffLines[i-1][0] == '-' {
					offsets, _ := diffIndexRange(line[1:], diffLines[i-1][1:])
					actual += breakWithColor(line, &c, offsets)
				} else {
					actual += breakWithColor(line, &c, []Range{{Start: 0, End: len(line) - 1}})
				}
			} else { // Process lines that do not start with '-' or '+'.
				expect += breakWithColor(line, nil, []Range{})
				actual += breakWithColor(line, nil, []Range{})
			}
		}
	}

	// Return the accumulated expected and actual strings.
	return expect, actual
}

// breakWithColor applies color to specific ranges within the input string and breaks the string into lines.
// input: The string to be processed.
// c: The color attribute to apply to the specified ranges. If nil, no color is applied.
// highlightRanges: A slice of Range structs specifying the start and end indices for color application.
func breakWithColor(input string, c *color.Attribute, highlightRanges []Range) string {
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
		if lineLen == MAX_LINE_LENGTH {
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

const MAX_LINE_LENGTH = 50

// breakLines breaks the input string into lines of a specified maximum length.
// input: The string to be processed and broken into lines.
// Returns the input string with line breaks inserted at the specified maximum length.
func breakLines(input string) string {
	var output strings.Builder      // Builder for the resulting output string.
	var currentLine strings.Builder // Builder for the current line being processed.
	inANSISequence := false         // Boolean to track if we are inside an ANSI escape sequence.
	lineLength := 0                 // Counter for the current line length.

	var ansiSequenceBuilder strings.Builder // Builder for the ANSI escape sequence.

	// Iterate over each character in the input string.
	for _, char := range input {
		if char == '\x1b' { // Check if the character is the start of an ANSI escape sequence.
			inANSISequence = true
		}
		if inANSISequence {
			ansiSequenceBuilder.WriteRune(char) // Add the character to the ANSI sequence builder.
			if char == 'm' {                    // Check if the ANSI escape sequence has ended.
				inANSISequence = false
				currentLine.WriteString(ansiSequenceBuilder.String()) // Add the completed ANSI sequence to the current line.
				ansiSequenceBuilder.Reset()                           // Reset the ANSI sequence builder.
			}
		} else {
			if isControlCharacter(char) && char != '\n' {
				currentLine.WriteRune(char) // Add control characters directly to the current line.
			} else {
				if lineLength >= MAX_LINE_LENGTH {
					output.WriteString(currentLine.String()) // Add the current line to the output.
					output.WriteRune('\n')                   // Add a newline character.
					currentLine.Reset()                      // Reset the current line builder.
					lineLength = 0                           // Reset the line length counter.
				} else if char == '\n' {
					output.WriteString(currentLine.String()) // Add the current line to the output.
					output.WriteRune(char)                   // Add the newline character.
					currentLine.Reset()                      // Reset the current line builder.
					lineLength = 0                           // Reset the line length counter.
				} else {
					currentLine.WriteRune(char) // Add the character to the current line.
					lineLength++                // Increment the line length counter.
				}
			}
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

// WrapTextWithAnsi processes the input string to ensure ANSI escape sequences are properly wrapped across lines.
// input: The input string containing text and ANSI escape sequences.
// Returns the processed string with properly wrapped ANSI escape sequences.
func wrapTextWithAnsi(input string) string {
	scanner := bufio.NewScanner(strings.NewReader(input)) // Create a scanner to read the input string line by line.
	var wrappedBuilder strings.Builder                    // Builder for the resulting wrapped text.
	currentAnsiCode := ""                                 // Variable to hold the current ANSI escape sequence.
	lastAnsiCode := ""                                    // Variable to hold the last ANSI escape sequence.

	// Iterate over each line in the input string.
	for scanner.Scan() {
		line := scanner.Text() // Get the current line.

		// If there is a current ANSI code, append it to the builder.
		if currentAnsiCode != "" {
			wrappedBuilder.WriteString(currentAnsiCode)
		}

		// Find all ANSI escape sequences in the current line.
		startAnsiCodes := ansiRegex.FindAllString(line, -1)
		if len(startAnsiCodes) > 0 {
			// Update the last ANSI escape sequence to the last one found in the line.
			lastAnsiCode = startAnsiCodes[len(startAnsiCodes)-1]
		}

		// Append the current line to the builder.
		wrappedBuilder.WriteString(line)

		// Check if the current ANSI code needs to be reset or updated.
		if (currentAnsiCode != "" && !strings.HasSuffix(line, ansiResetCode)) || len(startAnsiCodes) > 0 {
			// If the current line does not end with a reset code or if there are ANSI codes, append a reset code.
			wrappedBuilder.WriteString(ansiResetCode)
			// Update the current ANSI code to the last one found in the line.
			currentAnsiCode = lastAnsiCode
		} else {
			// If no ANSI codes need to be maintained, reset the current ANSI code.
			currentAnsiCode = ""
		}

		// Append a newline character to the builder.
		wrappedBuilder.WriteString("\n")
	}

	// Return the processed string with properly wrapped ANSI escape sequences.
	return wrappedBuilder.String()
}

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

// SprintDiffHeader compares the headers of the expected and actual maps and returns the differences as colorized strings.
// expect: The map containing the expected header values.
// actual: The map containing the actual header values.
// Returns two strings: the colorized differences for the expected and actual headers.
func SprintDiffHeader(expect, actual map[string]string) (string, string) {
	var expectAll, actualAll strings.Builder // Builders for the resulting strings.

	// Iterate over each key-value pair in the expected map.
	for key, expValue := range expect {
		actValue := actual[key] // Get the corresponding value from the actual map.

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
	return expectAll.String(), actualAll.String()
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
		} else {
			// If it isn't, append the word as-is to the result.
			result.WriteString(word + " ")
		}
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
// It returns a slice of Range structs indicating the start and end indices of differences and a boolean indicating if there are differences.
func diffIndexRange(s1, s2 string) ([]Range, bool) {
	var ranges []Range // Slice to hold the ranges of differences.
	diff := false      // Boolean to track if there are any differences.

	// Split the input strings into slices of words.
	words1 := strings.Split(s1, " ")
	words2 := strings.Split(s2, " ")

	// Determine the maximum length between the two word slices.
	maxLen := len(words1)
	if len(words2) > maxLen {
		maxLen = len(words2)
	}

	startIndex := 0 // Initialize the starting index for the ranges.

	// Iterate over the words up to the maximum length.
	for i := 0; i < maxLen; i++ {
		word1, word2 := "", "" // Initialize variables for the current words in each string.

		// Assign the current word from the first string if within bounds.
		if i < len(words1) {
			word1 = words1[i]
		}
		// Assign the current word from the second string if within bounds.
		if i < len(words2) {
			word2 = words2[i]
		}

		// If the words at the current index are different, record the range.
		if word1 != word2 {
			if !diff {
				diff = true // Set the difference flag to true.
			}
			// Calculate the end index for the current range.
			endIndex := startIndex + len(word1)
			// Append the range of the differing words to the ranges slice.
			ranges = append(ranges, Range{Start: startIndex, End: endIndex - 1})
		}
		// Update the starting index for the next word.
		startIndex += len(word1) + 1
	}

	// Return the ranges of differences and the difference flag.
	return ranges, diff
}

// diffArrayRange calculates the indices of differences between two strings of words.
// It returns the indices where the words differ in both strings, and a boolean indicating if there are differences.
func diffArrayRange(s1, s2 string) ([]int, []int, bool) {
	var indices1, indices2 []int // Slices to hold the indices of differences for each string.
	diff := false                // Boolean to track if there are any differences.

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
		word1, word2 := "", "" // Initialize variables for the current words in each string.

		// Assign the current word from the first string if within bounds.
		if i < len(words1) {
			word1 = words1[i]
		}
		// Assign the current word from the second string if within bounds.
		if i < len(words2) {
			word2 = words2[i]
		}

		// If the words at the current index are different, record the index.
		if word1 != word2 {
			// Add the index to the first string's differences if within bounds.
			if i < len(words1) {
				indices1 = append(indices1, i)
			}
			// Add the index to the second string's differences if within bounds.
			if i < len(words2) {
				indices2 = append(indices2, i)
			}
			diff = true // Set the difference flag to true.
		}
	}

	// Handle the case where the first string is longer than the second string.
	if len(words1) > len(words2) {
		for i := len(words2); i < len(words1); i++ {
			indices1 = append(indices1, i) // Record the remaining indices as differences.
		}
		// Handle the case where the second string is longer than the first string.
	} else if len(words2) > len(words1) {
		for i := len(words1); i < len(words2); i++ {
			indices2 = append(indices2, i) // Record the remaining indices as differences.
		}
	}

	// Return the indices of differences for both strings and the difference flag.
	return indices1, indices2, diff
}
