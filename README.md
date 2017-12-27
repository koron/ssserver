# Web screenshot server

Start the server, it listen 3000 on localhost.

    $ go run

Request from other console

    $ curl -v http://127.0.0.1:3000/https://www.kaoriya.net/ -o kaoriya.png
    $ curl -v http://127.0.0.1:3000/http://tokyo-ame.jwa.or.jp/ -o ame.png
