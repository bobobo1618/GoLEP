#GoLEP

This is a basic utility to serve [Lepton](https://github.com/dropbox/lepton) files over HTTP.

```
Usage of ./golep:
  -leptonSocket string
    	Socket to use to connect to Lepton. e.g. tcp://localhost:2402, unix:///tmp/.leptonsock (default "tcp://localhost:2402")
  -listen string
    	Host and port to listen on. (default ":8080")
  -readTimeout int
    	Timeout for reading. (default 10)
  -root string
    	Root path to serve files from. (default ".")
  -writeTimeout int
    	Timeout for writing. (default 10)
````
