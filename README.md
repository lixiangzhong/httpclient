# httpclient
go(golang) http request


###Example Get
```go
package main

import (
	"fmt"
	"github.com/lixiangzhong/httpclient"
	"time"
)

func main() {
	c := httpclient.New()
	c.Get("www.google.com")//if url Scheme=="" default:http://www.google.com
	c.SetTimeout(3 * time.Second)
	c.Query.Add("key", "value")
	res, err := c.Do()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.StatusCode)
}
```

###Example Post
```go
package main

import (
	"fmt"
	"github.com/lixiangzhong/httpclient"
)

func main() {
	c := httpclient.New()
	c.PostForm("www.google.com/example/api")
	c.Param.Add("key", "value")
	res, err := c.Do()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.String())
}
```
