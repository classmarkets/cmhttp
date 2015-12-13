# Composable HTTP client decorators


## USAGE

    baseClient := &http.Client{
        Transport: &http.Transport{},
    }
    cmhttp.MustConfigureTLS(baseClient.Transport.(*http.Transport), "/etc/ssl/cabundle.pem")

    client := cmhttp.Decorate(
        baseClient,
        cmhttp.TokenAuthenticated("my little secret"),
        cmhttp.Scoped("https://api.classmarkets.com"),
        cmhttp.Typed("application/json"), // or cmhttp.JSON()
    )

    req, err := http.NewRequest("GET", "geo/v1/area/de.berlin", nil)
    if err != nil {
        panic(err)
    }

    resp, err := client.Do(req)

Implementing custom decorators:

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
                        r.Method,
                        r.URL,
                        resp.StatusCode,
                        err,
                        time.Since(begin)/1e6,
                    )
                }(time.Now())

                resp, err = c.Do(r)
                return resp, err
            })
        }   
    }

Use it:

    logger := log.New(os.Stdout, "my-service: ", log.LstdFlags)
    client := Logged(logger)(http.DefaultClient)
