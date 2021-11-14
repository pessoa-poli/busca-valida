package esqueries

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	//"net"

	mo "buscavalida/models"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	esapi "github.com/elastic/go-elasticsearch/v7/esapi"
)

var (
	ES_CLIENT *elasticsearch.Client
)

//InitES ... Inits the global ES_CLIENT
func InitES(ElasticIP, ElasticPort, ElasticUserName, ElasticPassword string) {
	ES_CLIENT = esCreateClient(ElasticIP, ElasticPort, ElasticUserName, ElasticPassword)
}

//createEsClient ... Creates a elasticsearchClient base on environment variables defined into the docker-compose.yml file.
func esCreateClient(ElasticIP, ElasticPort, ElasticUserName, ElasticPassword string) (es *elasticsearch.Client) {
	elasticURL := fmt.Sprintf("http://%v:%v", ElasticIP, ElasticPort)
	cfg := elasticsearch.Config{
		Addresses: []string{
			elasticURL,
		},
		Username: ElasticUserName,
		Password: ElasticPassword}
	es, _ = elasticsearch.NewClient(cfg)
	return es
}

//esGetTickets ... Gets the tickets from the elastic DB using a PIT for scrolling over results.
func esGetTickets(cruIndex, sipCallID, fromTag, toTag, keepAlive string) (myTickets mo.DocumentStruct) {
	fmt.Println("Getting tickets.")
	//Define an empty search_after string
	searchAfter := ""
	//Open a PIT for the specified index
	pit := esOpenPIT(cruIndex, keepAlive)
	//Build the request, creating a function here, so we can rebuild it for each loop, with the right PIT.
	buildBody := func() string {
		req := fmt.Sprintf(`{
		"size": 100,
		"query": {
		  "bool": {
			"filter": [
			  {
				"term": {
				  "sipCallId.keyword": "%s"
				}
			  },			  
			  {
				"term": {
				  "fromTag.keyword": "%s"
				}
			  },
			  {
				"term": {
				  "toTag.keyword": "%s"
				}
			  }
			]
		  }
		},
		"pit": {
		  "id": "%s",
		  "keep_alive": "%s"
		},
		"sort": [ 
    {"_id": "asc"}
  ]%s
	  }`, sipCallID, fromTag, toTag, pit, keepAlive, searchAfter)
		return req
	}

	for {
		//Build the request body
		requestBody := buildBody()
		/* if globalDebug {
			fmt.Println("RequestBody: " + requestBody)
		} */

		//Build the request
		req := esapi.SearchRequest{
			Index: []string{},
			Body:  strings.NewReader(requestBody),
		}

		//Create a var to hold the midway structs to parse the request
		var midTickets mo.DocumentStruct

		//DO the request
		resp, err := req.Do(context.Background(), ES_CLIENT)
		if err != nil {
			if resp.Body != nil {
				respFull, _ := ioutil.ReadAll(resp.Body)
				fmt.Printf("\nRequest response: %s\n", respFull)
			}
			panic(err.Error())
		}
		fmt.Printf("\nResp Status: %s\n", resp.Status())
		//Read the request response
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Failed to read response.")
			panic(err.Error())
		}
		//Print the response for debbuging
		fmt.Printf("\nResponse is: %s\n", bodyBytes)
		//Marshal it in a struct
		err = json.Unmarshal(bodyBytes, &midTickets)
		if err != nil {
			fmt.Println("Failed to unmarshall response into tickets")
			panic(err.Error())
		}
		fmt.Printf("Found %v tickets\n", len(midTickets.Hits.Hits))
		if len(midTickets.Hits.Hits) == 0 {
			goto Exit
		}
		//Append results to the list of tickets
		myTickets.Hits.Hits = append(myTickets.Hits.Hits, midTickets.Hits.Hits...)
		//Get the new PIT to use on the next request for tickets.
		var respInterface map[string]interface{}
		err = json.Unmarshal(bodyBytes, &respInterface)
		if err != nil {
			panic(err.Error())
		}
		pit = respInterface["pit_id"].(string)
		//Get the new search_after value from the sort field of the last response hit
		searchAfterList := respInterface["hits"].(map[string]interface{})["hits"].([]interface{})
		searchAfterValue := searchAfterList[len(searchAfterList)-1].(map[string]interface{})["sort"].([]interface{})[0].(string)
		searchAfter = fmt.Sprintf(`,"search_after":["%s"]`, searchAfterValue)
		fmt.Printf("\nSearch after value is: %s\n", searchAfterValue)
	}
Exit:
	fmt.Printf("\nTotal tickets collected: %v\n", len(myTickets.Hits.Hits))
	return myTickets
}

//esOpenPIT ... Opens a point in time which is the way Pagination is done in ElasticSearch. This function is used by esGetTickets
func esOpenPIT(index, keepalive string) (pit string) {
	req := esapi.OpenPointInTimeRequest{
		Index:     []string{index},
		KeepAlive: keepalive,
	}

	resp, err := req.Do(context.Background(), ES_CLIENT)
	if err != nil {
		panic(err.Error())
	}
	respFull, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	var pitInterface map[string]interface{}
	err = json.Unmarshal(respFull, &pitInterface)
	if err != nil {
		panic(err.Error())
	}
	pit = pitInterface["id"].(string)
	//log.Printf("PIT: " + pit)
	return pit
}
