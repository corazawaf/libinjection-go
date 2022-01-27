# libinjection
libinjection is a Go porting of the libinjection([http://www.client9.com/projects/libinjection/](http://www.client9.com/projects/libinjection/)) and it's thread safe.

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
package main

import (
	"fmt"
	"github.com/bxlxx/libinjection-go"
)

func main() {
	fmt.Println("result: ", libinjection.IsXSS("<script>alert('1')</script>"))
}
```

## License
libinjection-golang is distributed under the same license as the [libinjection](http://www.client9.com/projects/libinjection/).