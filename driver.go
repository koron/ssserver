package main

import (
	"fmt"
	"runtime"

	"github.com/sclevine/agouti"
)

var driverName string

func newWebDriver() (*agouti.WebDriver, error) {
	switch driverName {
	case "chrome":
		return chromeDriver(), nil
	case "firefox":
		return firefoxDriver(), nil
	default:
		return nil, fmt.Errorf("unknown web driver: %q", driverName)
	}
}

var chromeArgs = []string{
	"headless",
	"disable-gpu",
	"hide-scrollbars",
}

func chromeDriver() *agouti.WebDriver {
	infof("using chrome driver")
	return agouti.ChromeDriver(agouti.ChromeOptions("args", chromeArgs))
}

var firefoxArgs = []string{
	"-headless",
}

func firefoxDriver() *agouti.WebDriver {
	infof("using firefox driver")
	return FirefoxDiver(agouti.Desired(agouti.Capabilities{
		"moz:firefoxOptions": map[string]interface{}{
			"args": firefoxArgs,
		},
	}))
}

// FirefoxDiver creates a WebDriver of Firefox.
func FirefoxDiver(options ...agouti.Option) *agouti.WebDriver {
	var binaryName string
	if runtime.GOOS == "windows" {
		binaryName = "geckodriver.exe"
	} else {
		binaryName = "geckodriver"
	}
	command := []string{binaryName, "--port={{.Port}}"}
	return agouti.NewWebDriver("http://{{.Address}}", command, options...)
}
