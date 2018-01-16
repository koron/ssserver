# Web screenshot server

Install the command.

    $ go get -u github.com/koron/ssserver

Start the server, it listen 3000 on localhost.

    $ ssserver

Request from other terminal or console

    $ curl -v http://127.0.0.1:3000/?u=https://www.kaoriya.net/ -o kaoriya.png
    $ curl -v http://127.0.0.1:3000/?u=http://tokyo-ame.jwa.or.jp/ -o ame.png

## Requirements

*   [Chrome browser][browser]
*   [Chrome driver][driver]

    Install the driver to anywhere in your PATH.

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

## Examples

Full page screen shot:

    $ curl -v 'http://127.0.0.1:3000/?u=https://www.kaoriya.net/&full' -o kaoriya-full.png

[browser]:https://www.google.com/chrome/browser/desktop/index.html
[driver]:https://sites.google.com/a/chromium.org/chromedriver/downloads
[dur]:https://golang.org/pkg/time/#ParseDuration
