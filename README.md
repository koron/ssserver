# Web screenshot server

[![CircleCI](https://circleci.com/gh/koron/ssserver.svg?style=svg)](https://circleci.com/gh/koron/ssserver)
[![Go Report Card](https://goreportcard.com/badge/github.com/koron/ssserver)](https://goreportcard.com/report/github.com/koron/ssserver)

Install the command.

    $ go get -u github.com/koron/ssserver

Assure to install Chrome and Chrome driver.

Then start the server, it listen 3000 on localhost.

    $ ssserver

Run these command from other terminal or console to obtain screenshots.

    $ curl -v http://127.0.0.1:3000/?u=https://www.kaoriya.net/ -o kaoriya.png
    $ curl -v http://127.0.0.1:3000/?u=http://tokyo-ame.jwa.or.jp/ -o ame.png

## Requirements

These tools are required to make ssserver work correctly.

*   for Chrome
    *   [Chrome browser][chrome_browser]
    *   [Chrome driver][chrome_driver]
*   for Firefox (WIP: with many problems)
    *   [Firefox driver][firefox_driver]

Install those to anywhere in you PATH environment variable.

## Request Parameters

Name        |Description
------------|---------------------------------------------------------------
`u`         |URL to take a screenshot (mandatory)
`w`         |Width of screenshot (default: 1024)
`h`         |Height of screenshot (default: 768)
`wait`      |Wait before take a screenshot (default: 0, see [Duration][dur])
`sX`        |Scroll offset X (default: 0)
`sY`        |Scroll offset Y (default: 0)
`full`      |Full page screenshot. Ignore `h`, `sX` and `sY` when set.
`save`      |Show save as dialog in browser.  Imply `Content-Disposition: attachment`

## Server Options

Name        |Description
------------|---------------------------------------------------------------
`-addr`     |Server listen address (default: ":3000")
`-driver`   |WebDriver name (default: "chrome")
`-maxpages` |Max number of browser instances (default: 4)
`-v`        |Verbose logging

## Examples

Full page screenshot:

    $ curl -v 'http://127.0.0.1:3000/?u=https://www.kaoriya.net/&full' -o kaoriya-full.png

[chrome_browser]:https://www.google.com/chrome/browser/desktop/index.html
[chrome_driver]:https://sites.google.com/a/chromium.org/chromedriver/downloads
[firefox_driver]:https://github.com/mozilla/geckodriver/releases
[dur]:https://golang.org/pkg/time/#ParseDuration
