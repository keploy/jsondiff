<p align="center">
  <img align="center" src="https://docs.keploy.io/img/keploy-logo-dark.svg?s=200&v=4" height="40%" width="40%" alt="keploy logo"/>
</p>
<h3 align="center">
<b>
⚡️ API tests faster than unit tests, from user traffic ⚡️
</b>
</h3>
<p align="center">
🌟 The must-have tool for developers in the AI-Gen era 🌟
</p>

---

<h4 align="center">

<a href="https://twitter.com/Keploy_io">
    <img src="https://img.shields.io/badge/follow-%40keployio-1DA1F2?logo=twitter&style=social" alt="Keploy Twitter" />
  </a>

<a href="https://github.com/Keploy/Keploy/issues">
    <img src="https://img.shields.io/github/stars/keploy/keploy?color=%23EAC54F&logo=github&label=Help us reach 4k stars! Now at:" alt="Help us reach 4k stars!" />
  </a>

  <a href="https://landscape.cncf.io/?item=app-definition-and-development--continuous-integration-delivery--keploy">
    <img src="https://img.shields.io/badge/CNCF%20Landscape-5699C6?logo=cncf&style=social" alt="Keploy CNCF Landscape" />
  </a>

[![Slack](https://img.shields.io/badge/Slack-4A154B?style=for-the-badge&logo=slack&logoColor=white)](https://join.slack.com/t/keploy/shared_invite/zt-2dno1yetd-Ec3el~tTwHYIHgGI0jPe7A)
[![LinkedIn](https://img.shields.io/badge/linkedin-%230077B5.svg?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/company/keploy/)
[![YouTube](https://img.shields.io/badge/YouTube-%23FF0000.svg?style=for-the-badge&logo=YouTube&logoColor=white)](https://www.youtube.com/channel/UC6OTg7F4o0WkmNtSoob34lg)
[![Twitter](https://img.shields.io/badge/Twitter-%231DA1F2.svg?style=for-the-badge&logo=Twitter&logoColor=white)](https://twitter.com/Keployio)

</h4>

# jsonDiff
`jsonDiff` is a Go package that allows you to compare JSON objects and return the differences as colorized strings.

## Features
- Compare two JSON objects and highlight the differences.
- Supports comparing headers of expected and actual maps.
- Provides colorized differences for easy visualization.

## Installation

```sh
go get github.com/keploy/jsonDiff
```

# Usage

## Comparing JSON Objects

```sh
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	jsonDiff "github.com/keploy/jsonDiff"
	"github.com/olekukonko/tablewriter"
)

func main() {
	json1 := []byte("{\"animals\":[{\"name\":\"Dog\"},{\"name\":\"Cat\"},{\"name\":\"Elephant\"}]}")
	json2 := []byte("{\"animals\":[{\"name\":\"Dog\"},{\"name\":\"Cat\"},{\"apple\":\"lusiancs\"},{\"name\":\"Elephant\"}]}")

	diff, err := jsonDiff.CompareJSON(json1, json2, nil, false)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	result := expectActualTable(diff.Actual, diff.Expected, "", false)
	println(result)
}

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

func expectActualTable(exp string, act string, field string, centerize bool) string {
	buf := &bytes.Buffer{}
	table := tablewriter.NewWriter(buf)

	if centerize {
		table.SetAlignment(tablewriter.ALIGN_CENTER)
	} else {
		table.SetAlignment(tablewriter.ALIGN_LEFT)
	}
	// jsonDiff.JsonDiff()
	exp = wrapTextWithAnsi(exp)
	act = wrapTextWithAnsi(act)
	table.SetHeader([]string{fmt.Sprintf("Expect %v", field), fmt.Sprintf("Actual %v", field)})
	table.SetAutoWrapText(false)
	table.SetBorder(false)
	table.SetColMinWidth(0, maxLineLength)
	table.SetColMinWidth(1, maxLineLength)
	table.Append([]string{exp, act})
	table.Render()
	return buf.String()
}

const maxLineLength = 50

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

var ansiResetCode = "\x1b[0m"
```

## Comparing Headers

```sh
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	jsonDiff "github.com/keploy/jsonDiff"
	"github.com/olekukonko/tablewriter"
)

func main() {
	json1 := map[string]string{
		"Etag": "W/\"1c0-4VkjzPwyKEH0Xy9lGO28f/cyPk4\"",
		"Vary": "a",
	}
	json2 := map[string]string{
		"Etag": "W/\"1c0-8j/k9MOCbWGtKgVesjFGmY6dEAs\"",
		"Vary": "Origin",
	}
	diff := jsonDiff.CompareHeaders(json1, json2)

	result := expectActualTable(diff.Actual, diff.Expected, "", false)
	println(result)
}

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

func expectActualTable(exp string, act string, field string, centerize bool) string {
	buf := &bytes.Buffer{}
	table := tablewriter.NewWriter(buf)

	if centerize {
		table.SetAlignment(tablewriter.ALIGN_CENTER)
	} else {
		table.SetAlignment(tablewriter.ALIGN_LEFT)
	}
	// jsonDiff.JsonDiff()
	exp = wrapTextWithAnsi(exp)
	act = wrapTextWithAnsi(act)
	table.SetHeader([]string{fmt.Sprintf("Expect %v", field), fmt.Sprintf("Actual %v", field)})
	table.SetAutoWrapText(false)
	table.SetBorder(false)
	table.SetColMinWidth(0, maxLineLength)
	table.SetColMinWidth(1, maxLineLength)
	table.Append([]string{exp, act})
	table.Render()
	return buf.String()
}

const maxLineLength = 50

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

var ansiResetCode = "\x1b[0m"

```

## 👨🏻‍💻 Let's Build Together! 👩🏻‍💻
Whether you're a newbie coder or a wizard 🧙‍♀️, your perspective is golden. Take a peek at our:

📜 [Contribution Guidelines](https://github.com/keploy/JsonDiff/blob/main/CONTRIBUTING.md)

❤️ [Code of Conduct](https://github.com/keploy/keploy/blob/main/CODE_OF_CONDUCT.md)
