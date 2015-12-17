# Composable HTTP client decorators

### Simple Usage

```golang
client := cmhttp.Decorate(
    http.DefaultClient,
    cmhttp.Scoped("https://api.example.com"),
    cmhttp.Typed("application/json"), // or cmhttp.JSON()
)

req, err := http.NewRequest("GET", "/v1/places/de.berlin", nil)
if err != nil {
    panic(err)
}

resp, err := client.Do(req)
```

### Configuring TLS

```golang
baseClient := &http.Client{
    Transport: &http.Transport{},
}
cmhttp.MustConfigureTLS(baseClient.Transport.(*http.Transport), "/etc/ssl/cabundle.pem")

client := cmhttp.Decorate(
    baseClient,
    cmhttp.Scoped("https://api.example.com"),
    cmhttp.Typed("application/json"), // or cmhttp.JSON()
)
```

### Implementing custom decorators:

```golang
import (
    "log"
    "net/http"
    "os"
    "time"
)

func Logged(log *log.Logger) Decorator {
    return func(c Client) Client {
        return ClientFunc(func(r *http.Request) (*http.Response, error) {
            var (
                resp *http.Response
                err error
            )

            defer func(begin time.Time) {
                log.Printf(
                    "method=%s url=%s resp=%d err=%s took=%dms",
                    r.Method, r.URL, resp.StatusCode, err, time.Since(begin)/1e6,
                )
            }(time.Now())

            resp, err = c.Do(r)
            return resp, err
        })
    }   
}
```

Use it:

    logger := log.New(os.Stdout, "my-service: ", log.LstdFlags)
    client := Logged(logger)(http.DefaultClient)
