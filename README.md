# Req 
CLI that makes a request to a server and displays the body of the response.
Support this protocos:
- HTTP
- GraphQL
- Websocket

# Install
TODO

Using go 


# Usage 

### HTTP 

GET
```bash
$ req -u http://localhost:8080
```

POST - send JSON body payload
```bash
$ req -m post -u https://jsonplaceholder.typicode.com/posts -p '{"title": "foo", "body": "bar", "userId": 1}'
```

POST - send form data
```bash
$ req -m post  -u https://site.com -p "foo=bar&jhon=doe"
```

### Websocket

Listen for messages
```bash
$ req -t ws -u wss://socketsbay.com/wss/v2/1/demo/
```

Send message
```bash
$ req -t ws -p "some message"  -u wss://socketsbay.com/wss/v2/1/demo/
```


### Graphql

Send query to graphql server
```bash 
$ req . -t gq -u https://countries.trevorblades.com/ -p 'query {countries {name}}'
```

# Parameters

| Parameter | Description |
| --- | --- |
| -u | url server |
| -m | http method |
| -p | data to send to server  |
| -t | type protocol (http, ws, gq) defaut http |
| -q | http query params |
| -h | http headers |
| -v | show server header response |
