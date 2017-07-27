# A GO Client for SMAP

This code was written to be a GO based client for [SMAP](https://github.com/SoftwareDefinedBuildings/smap).
The code is not fully implemted, and at the time of writing, and only supports subscribtions.

## Usage

Here is an exaple of a subscribtion:

```go
output := make(chan smap.SubscribtionMessage, 1000)
quit := make(chan bool, 1)

client := smap.NewClient("http://www.someurl.com:8079")
client.Subscribe(output, quit, "Metadata/SourceName='SomeKey'")

go func() {
    time.Sleep(time.Second * 50)
    quit <- true
}()

for item := range output {
    fmt.Println(item.Path)
}
```
