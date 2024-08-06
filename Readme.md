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
	"fmt"
	"github.com/keploy/jsonDiff"
)

func main() {
	json1 := []byte(`{"name": "Alice", "age": 30, "city": "New York"}`)
	json2 := []byte(`{"name": "Alice", "age": 31, "city": "Los Angeles"}`)

	diff, err := jsonDiff.ColorJSONDiff(json1, json2, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Expected Response:")
	fmt.Println(diff.ExpectedResponse)
	fmt.Println("Actual Response:")
	fmt.Println(diff.ActualResponse)
}
```

## Comparing Headers

```sh
package main

import (
	"fmt"
	"github.com/yourusername/jsonDiff"
)

func main() {
	expectedHeaders := map[string]string{
		"Content-Type": "application/json",
		"Authorization": "Bearer token123",
	}

	actualHeaders := map[string]string{
		"Content-Type": "application/json",
		"Authorization": "Bearer token456",
	}

	diff := jsonDiff.ColorHeaderDiff(expectedHeaders, actualHeaders)

	fmt.Println("Expected Headers:")
	fmt.Println(diff.ExpectedResponse)
	fmt.Println("Actual Headers:")
	fmt.Println(diff.ActualResponse)
}

```

## 👨🏻‍💻 Let's Build Together! 👩🏻‍💻
Whether you're a newbie coder or a wizard 🧙‍♀️, your perspective is golden. Take a peek at our:

📜 [Contribution Guidelines](https://github.com/keploy/JsonDiff/blob/main/CONTRIBUTING.md)

❤️ [Code of Conduct](https://github.com/keploy/keploy/blob/main/CODE_OF_CONDUCT.md)
