# Web screenshot server

Start the server, it listen 3000 on localhost.

    $ go run

Request from other console

    $ curl -v http://127.0.0.1:3000/?u=https://www.kaoriya.net/ -o kaoriya.png
    $ curl -v http://127.0.0.1:3000/?u=http://tokyo-ame.jwa.or.jp/ -o ame.png

## Request Parameters

Name        |Description
------------|---------------------------------------------------------------
`u`         |URL to take a screenshot (mandatory)
`w`         |Width of screenshot (default: 1024)
`h`         |Height of screenshot (default: 768)
`wait`      |Wait before take a screenshot (default: 0, see [Duration][dur])

[dur]:https://golang.org/pkg/time/#ParseDuration
