# jumpcloud interview take-home exercise (in Go)

# Instructions

Clone the repo:
```
git clone https://github.com/scruffycatherder/jumpcloud
```

Build and start the server:
```
go build
./jumpcloud --port=\<port>
```

Request a new hashed password:
```
curl —data “password=<password>” http://<hostname>:<port>/hash
(Returns the new hash_key)
```

Check for a completed hash:
```
curl —data “password=<password>” http://<hostname>:<port>/hash/<hash_key>
```

Request execution statistics:
```
curl —data “password=<password>” http://<hostname>:<port>/stats/
```

# Requirements (copied from provided doc for convenience)

Hash and Encode a Password String
Write an HTTP server that listens on a given port. Your server should be able to process multiple connections simultaneously. And provide the following endpoints:

Accept POST requests on the /hash endpoint with a form field named password to provide the value to hash. An incrementing identifier should return immediately but the password should not be hashed for 5 seconds. The hash should be computed as base64 encoded string of the SHA512 hash of the provided password.

For example, the first request to: 
curl —data “password=angryMonkey” http://localhost:8080/hash 
should return 1 immediately. The 42 request should return 42 immediately.  

5 seconds after the POST to /hash that returned an identifer you should be able to curl http://localhost:8080/hash/42 and get back the value of “ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==”
Statistics End-Point
Provide a statistics endpoint to get basic information about your password hashes. 

A GET request to /stats should return a JSON object with 2 key/value pairs. The “total” key should have a value for the count of POST requests to the /hash endpoint made to the server so far. The “average” key should have a value for the average time it has taken to process all of those requests in microseconds.

For example: curl http://localhost:8080/stats should return something like:
{“total”: 1, “average”: 123}
Graceful Shutdown
Provide support for a “graceful shutdown request”. If a request is made to /shutdown the server should reject new requests. The program should wait for any pending/in-flight work to finish before exiting.


