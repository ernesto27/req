# Req 
CLI that makes a request to a server and displays the body of the response.
Support this protocos:
- HTTP
- GraphQL
- Websocket
- GRPC

# Install

Using go 
```bash
go install github.com/ernesto27/req@latest
```

Using brew
```bash
brew install ernesto27/tools/req
```


# Usage 

### HTTP 

GET
```bash
req -u http://example.com
```

POST - send JSON body payload
```bash
req -m post -u https://jsonplaceholder.typicode.com/posts -p '{"title": "foo", "body": "bar", "userId": 1}'
```

POST - send form data
```bash
req -m post  -u https://site.com -p "foo=bar&jhon=doe"
```

### Websocket

Listen for messages
```bash
req -t ws -u wss://socketsbay.com/wss/v2/1/demo/
```

Send message
```bash
req -t ws -p "some message"  -u wss://socketsbay.com/wss/v2/1/demo/
```


### Graphql

Send query to graphql server
```bash 
req -t gq -u https://countries.trevorblades.com/ -p 'query {countries {name}}'
```
Use file to send query
```bash
req -t gq -u https://countries.trevorblades.com/ -p @myfolder/query.txt
```


### GRPC

Send request to grpc server
```bash
req -t grpc -u localhost:50051 -import-path /pathprotofiles/helloworld -proto helloworld.proto -p '{"name": "ernesto"}' -method helloworld.Greeter.SayHello
```


# Parameters

| Parameter | Description |
| --- | --- |
| -u | url server |
| -m | http method |
| -p | data to send to server in raw string of use @myfolder/file to send from file  |
| -t | type protocol (http, ws, gq, grpc) defaut http |
| -q | http query params |
| -h | http headers |
| -v | show server header response |
| -d | Download response to file|
| -import-path | GRPC - path to proto files |
| -proto | GRPC - proto file name |
| -method | GRPC - method to call |



