# libinjection
libinjection is a Go porting of the libinjection([http://www.client9.com/projects/libinjection/](http://www.client9.com/projects/libinjection/)).

## How to use
### SQLi Example
```go
package main

import (
    "fmt"
    "github.com/bxlxx/libinjection-go"
)

func main() {
    result, fingerprint := libinjection.IsSQLi("-1' and 1=1 --")
    fmt.Println("=========result==========: ", result)
    fmt.Println("=======fingerprint=======: ", string(fingerprint[:]))
}
```

### XSS Example
```go

```

## Benchmark
### SQLi benchmark
```go

```

### XSS benchmark
```go

```

## License
libinjection-golang is distributed under the same license as libinjection.