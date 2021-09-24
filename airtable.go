package airtable

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
)

const (
	airtableKeyVariable  = "AIRTABLE_KEY"
	airtableBaseVariable = "AIRTABLE_BASE"
	airtableHostVariable = "AIRTABLE_HOST"
	defaultAirtableHost  = "https://api.airtable.com/v0/"
)

type AirTableClientInterface interface{}

type AirtableClient struct {
	Key    *string
	URL    *string
	Client http.Client
}

type AirtableRequest struct {
	Method  string
	Table   string
	Payload *AirtablePayload
	URL     *string
}

type AirtableRecord struct {
	CreatedTime string      `json:"createdTime,omitempty"`
	Fields      interface{} `json:"fields,omitempty"`
	Id          string      `json:"id,omitempty"`
}

type AirtablePayload struct {
	Records []AirtableRecord `json:"records,omitempty"`
}

func InitializeClient() (*AirtableClient, error) {

	// for glog and anything else
	flag.Parse()

	glog.Info("Starting airtable service")

	airtableUrl, err := generateAirtableURL()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to connect to generate AirTable URL %v", err))
	}

	airtableKey, isSet := os.LookupEnv(airtableKeyVariable)

	if !isSet || airtableKey == "" {
		return nil, errors.New("The AIRTABLE_KEY environment variable is not set")
	}

	return &AirtableClient{
		Key:    &airtableKey,
		URL:    airtableUrl,
		Client: initAirtableClient(),
	}, nil

}

func (c *AirtableClient) SendRequest(req *AirtableRequest) ([]byte, error) {
	// url := fmt.Sprintf(*req.URL)

	if req == nil {
		glog.Error("The request URL in SendRequest is nil")
		return nil, errors.New("The request URL in SendRequest is nil")
	}

	httpReq, err := req.buildHttpRequest(*req.URL, c.Key)
	if err != nil {
		glog.Errorf("Error sending the AirtableRequest %s", err)
		return nil, err
	}

	glog.Infof("Generated HTTP request %s", httpReq.Header.Get(authorizationHeader))
	glog.Infof("Sending request to %s using key %s", *req.URL, *c.Key)

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		glog.Errorf("Error sending request to airtable %s", err)
		return nil, err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		glog.Errorf("There was an error reading the response body %s", err)
	}

	glog.V(8).Infof("Got response body %v", string(respBytes))

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		glog.Errorf("Airtable returned an error status %d", resp.StatusCode)
		return nil, errors.New(
			fmt.Sprintf("Airtable response is an error %s", err))
	}

	return respBytes, nil
}

func (r *AirtableRequest) buildHttpRequest(url string, key *string) (*http.Request, error) {

	b := new(bytes.Buffer)

	err := json.NewEncoder(b).Encode(r.Payload)

	if err != nil {
		glog.Errorf("Encoded Bytes error %s", err)
		return nil, err
	}

	httpReq, err := http.NewRequest(
		r.Method,
		url,
		b)

	if err != nil {
		glog.Errorf("There was an error building the HTTP request %s", err)
		return nil, err
	}

	httpReq.Header.Add(contentHeader, jsonUtf8)
	httpReq.Header.Add(authorizationHeader, fmt.Sprintf(bearerString, *key))

	return httpReq, nil
}

func generateAirtableURL() (*string, error) {
	airtableBaseId, isSet := os.LookupEnv(airtableBaseVariable)

	if !isSet {
		return nil, errors.New("AirTable App ID is not set")
	}

	airtableHost, isSet := os.LookupEnv(airtableHostVariable)

	if !isSet {
		airtableHost = defaultAirtableHost
	}

	airtableBaseUrl := fmt.Sprintf(airtableHost+"%s", airtableBaseId) + "/%s"

	// url := fmt.Sprintf(airtableBaseUrl, airtableBaseVariable)

	glog.Infof("Initialized Airtable URL: %v", airtableBaseUrl)

	return &airtableBaseUrl, nil
}

func initAirtableClient() http.Client {
	return http.Client{
		Timeout: time.Second * 15,
	}
}

func (c *AirtableClient) MakeGetRecordRequest(table string, recordId string) *AirtableRequest {
	getRecordRequest := c.CreateAirtableRequest(http.MethodGet, table)
	*getRecordRequest.URL = fmt.Sprintf("%s/%s", *getRecordRequest.URL, recordId)
	glog.V(8).Infof("Updated client URL %s", *c.URL)
	return getRecordRequest
}

func (c *AirtableClient) MakeFilterRecordRequest(table string, filterQuery string) *AirtableRequest {
	filterRecordRequest := c.CreateAirtableRequest(http.MethodGet, table)
	*filterRecordRequest.URL = fmt.Sprintf("%s%s", *filterRecordRequest.URL, filterQuery)
	glog.V(8).Infof("Updated client URL for filter record %s", *c.URL)
	return filterRecordRequest
}

func (r *AirtableRequest) CreateRecord(fields interface{}) *AirtableRecord {

	glog.V(8).Infof("CreateRecord called with %+v", fields)

	airtableRecord := &AirtableRecord{
		Fields: &fields,
	}

	glog.V(8).Infof("Created airtableRecord %+v", airtableRecord)

	return airtableRecord
}

func (r *AirtableRequest) AddRecordToRequest(rec AirtableRecord) {

	r.Payload.Records = append(r.Payload.Records, rec)

	glog.V(8).Infof("AirtableRequest.Records after append %+v", r.Payload.Records)
}

func (c *AirtableClient) CreateAirtableRequest(method string, table string) *AirtableRequest {

	requestRecords := make([]AirtableRecord, 0)

	requestUrl := fmt.Sprintf(*c.URL, table)

	glog.V(8).Infof("requestUrl %s", requestUrl)
	airtableRequest := &AirtableRequest{
		Method:  method,
		Table:   table, // this can go away
		Payload: &AirtablePayload{Records: requestRecords},
		URL:     &requestUrl,
	}

	glog.V(8).Infof("Created airtableRequest %+v", airtableRequest)

	return airtableRequest
}
