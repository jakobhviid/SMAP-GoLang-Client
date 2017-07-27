package smap

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/buger/jsonparser"
)

//Client is a Client class that facilitates communication with the archiver
//Parseing implemented using: https://github.com/buger/jsonparser
type Client struct {
	url string
}

//NewClient creates a new instance of Client
//The url string must be in for format "http://www.somedomain.com:8079"
// usage example:
// output := make(chan Client.SubscribtionMessage, 1000)
// quit := make(chan bool, 1)
// client := Client.NewClient("http://URL:8079")
// client.Subscribe(output, quit, "Metadata/SourceName='SomeKey'")
// go func() {
// 	time.Sleep(time.Second * 10)
// 	quit <- true
// }()
// for item := range output {
// 	fmt.Println(item.Path)
// }
func NewClient(url string) Client {
	return Client{url: url}
}

//Subscribe starts a subscribtion (on republish) on the archiver and returns results in a channel
//An example for the subscribe filter is as following: "Metadata/SourceName='SomeKet'"
//A true message in the quitChannel quits the thread.
func (instance Client) Subscribe(outChannel chan SubscribtionMessage, quitChannel chan bool, subscribeFilter string) {

	go func() {
		postReader := bytes.NewReader([]byte(subscribeFilter))
		resp, err := http.Post(instance.url+"/republish", "application/json", postReader)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("Client posted and connected to " + instance.url + "/republish/ and subscribing to " + subscribeFilter)

		rdr := bufio.NewReader(resp.Body)

		for {
			select {
			//checks if the loop should be closed.
			case value, _ := <-quitChannel:
				if value {
					log.Println("closing output channel")
					close(outChannel)
					log.Println("Subscribtion loop quit for " + instance.url + " with filter " + subscribeFilter)
					return
				}
			//if not closing, keep reading.
			default:
				line, err := rdr.ReadSlice('\n')
				if err != nil {
					log.Println(err)
					return
				}
				// checking if the line is an empty line
				if string(line) != "\n" {
					//converting to JSON
					parsedLine := parseReading(line)
					//Outputting to channel
					for i := range parsedLine {
						outChannel <- parsedLine[i]
					}
				}
			}
		}
	}()
}

func parseReading(data []byte) []SubscribtionMessage {
	var x map[string]interface{} //used for parseing the keys
	keys := make([]string, 0)
	model := make([]SubscribtionMessage, 0) //actual model getting sent back

	//converts the array and puts it into x
	if err := json.Unmarshal(data, &x); err != nil {
		log.Println(err)
	} else {
		//Getting all keys first for parseing.
		for k := range x {
			keys = append(keys, k)
		}

		//Creating models
		for k := range keys {
			//getting data
			path := keys[k]
			uuid, _ := jsonparser.GetString(data, path, "uuid")
			readings := make([]SubscribtionReadingContainer, 0)

			//filling readings
			jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				//doing time first
				timeParseResult, _, _, _ := jsonparser.Get(value, "[0]")
				time, _ := strconv.ParseInt(string(timeParseResult), 10, 64)

				//doing the value
				valueParseResult, _, _, _ := jsonparser.Get(value, "[1]")
				rvalue := string(valueParseResult)

				readings = append(readings, SubscribtionReadingContainer{
					UnixTime: time,
					Value:    rvalue,
				})
			}, path, "Readings")

			//creating and appending mode
			model = append(model, SubscribtionMessage{
				Path:     path,
				UUID:     uuid,
				Readings: readings,
			})
		}
	}

	log.Printf("%v subscribtion messages were parsed.\n", len(model))

	return model
}
