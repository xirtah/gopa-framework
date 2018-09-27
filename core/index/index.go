package index

import (
	"encoding/json"

	"github.com/xirtah/gopa-framework/core/errors"
	log "github.com/xirtah/gopa-framework/core/logger/seelog"
	"github.com/xirtah/gopa-framework/core/model"
	"github.com/xirtah/gopa-framework/core/util"
)

// ElasticsearchConfig contains common settings for elasticsearch
type ElasticsearchConfig struct {
	Endpoint    string `config:"endpoint"`
	IndexPrefix string `config:"index_prefix"`
	Username    string `config:"username"`
	Password    string `config:"password"`
}

// ElasticsearchClient elasticsearch client api
type ElasticsearchClient struct {
	Config *ElasticsearchConfig
}

// InsertResponse is a index response object
type InsertResponse struct {
	Result  string `json:"result"`
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
}

// GetResponse is a get response object
type GetResponse struct {
	Found   bool                   `json:"found"`
	Index   string                 `json:"_index"`
	Type    string                 `json:"_type"`
	ID      string                 `json:"_id"`
	Version int                    `json:"_version"`
	Source  map[string]interface{} `json:"_source"`
}

// DeleteResponse is a delete response object
type DeleteResponse struct {
	Result  string `json:"result"`
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
}

// CountResponse is a count response object
type CountResponse struct {
	Count int `json:"count"`
}

// SearchResponse is a count response object
type SearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Hits     struct {
		Total    int                   `json:"total"`
		MaxScore float32               `json:"max_score"`
		Hits     []model.IndexDocument `json:"hits,omitempty"`
	} `json:"hits"`
	Aggregations map[string]model.Aggregation `json:"aggregations,omitempty"`
}

// RangeQuery is used to find value in range
type RangeQuery struct {
	Range map[string]map[string]interface{} `json:"range,omitempty"`
}

func (query *RangeQuery) Gt(field string, value interface{}) {
	query.Range = map[string]map[string]interface{}{}
	v := map[string]interface{}{}
	v["gt"] = value
	query.Range[field] = v
}

func (query *RangeQuery) Gte(field string, value interface{}) {
	query.Range = map[string]map[string]interface{}{}
	v := map[string]interface{}{}
	v["gte"] = value
	query.Range[field] = v
}

func (query *RangeQuery) Lt(field string, value interface{}) {
	query.Range = map[string]map[string]interface{}{}
	v := map[string]interface{}{}
	v["lt"] = value
	query.Range[field] = v
}

func (query *RangeQuery) Lte(field string, value interface{}) {
	query.Range = map[string]map[string]interface{}{}
	v := map[string]interface{}{}
	v["lte"] = value
	query.Range[field] = v
}

type MatchQuery struct {
	Match map[string]interface{} `json:"match,omitempty"`
}

type QueryStringQuery struct {
	Query map[string]interface{} `json:"query_string,omitempty"`
}

func NewQueryString(q string) *QueryStringQuery {
	query := QueryStringQuery{}
	query.Query = map[string]interface{}{}
	return &query
}

func (query *QueryStringQuery) QueryString(q string) {
	query.Query["query"] = q
}

func (query *QueryStringQuery) DefaultOperator(op string) {
	query.Query["default_operator"] = op
}

func (query *QueryStringQuery) Fields(fields ...string) {
	query.Query["default_operator"] = fields
}

// Init match query's condition
func (match *MatchQuery) Set(field string, v interface{}) {
	match.Match = map[string]interface{}{}
	match.Match[field] = v
}

// BoolQuery wrapper queries
type BoolQuery struct {
	Must    []interface{} `json:"must,omitempty"`
	MustNot []interface{} `json:"must_not,omitempty"`
	Should  []interface{} `json:"should,omitempty"`
}

// Query is the root query object
type Query struct {
	BoolQuery *BoolQuery `json:"bool"`
}

// SearchRequest is the root search query object
type SearchRequest struct {
	Query *Query         `json:"query,omitempty"`
	From  int            `json:"from"`
	Size  int            `json:"size"`
	Sort  *[]interface{} `json:"sort,omitempty"`
}

// AddSort add sort conditions to SearchRequest
func (request *SearchRequest) AddSort(field string, order string) {
	if (request.Sort) == nil {
		s := []interface{}{}
		request.Sort = &s
	}
	s := map[string]interface{}{}
	v := map[string]interface{}{}
	v["order"] = order
	s[field] = v
	*request.Sort = append(*request.Sort, s)
}

// IndexDoc index a document into elasticsearch - SAMEER review
func (c *ElasticsearchClient) Index(indexName, id string, data interface{}) (*InsertResponse, error) {
	if c.Config.IndexPrefix != "" {
		indexName = c.Config.IndexPrefix + indexName
	}
	url := c.Config.Endpoint + "/" + indexName + "/doc/" + id

	js, err := json.Marshal(data)

	//log.Info("indexing doc: ", url, ",", string(js))
	log.Debug("indexing doc: ", url, ",", string(js))

	if err != nil {
		return nil, err
	}
	//log.Info("1")
	req := util.NewPostRequest(url, js)
	//log.Info("2")
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	//log.Info("3")
	response, err := util.ExecuteRequest(req)
	//log.Info("4")
	if err != nil {
		log.Info("4e", err)
		return nil, err
	}
	//log.Info("5")

	//log.Info("indexing response: ", string(response.Body))
	log.Trace("indexing response: ", string(response.Body))

	esResp := &InsertResponse{}
	err = json.Unmarshal(response.Body, esResp)
	if err != nil {
		return &InsertResponse{}, err
	}
	if !(esResp.Result == "created" || esResp.Result == "updated") {
		return nil, errors.New(string(response.Body))
	}

	return esResp, nil
}

// Get fetch document by id
func (c *ElasticsearchClient) Get(indexName, id string) (*GetResponse, error) {
	if c.Config.IndexPrefix != "" {
		indexName = c.Config.IndexPrefix + indexName
	}
	url := c.Config.Endpoint + "/" + indexName + "/doc/" + id

	log.Debug("get doc: ", url)

	req := util.NewGetRequest(url)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	response, err := util.ExecuteRequest(req)

	if err != nil {
		return nil, err
	}

	log.Trace("get response: ", string(response.Body))

	esResp := &GetResponse{}
	err = json.Unmarshal(response.Body, esResp)
	if err != nil {
		return &GetResponse{}, err
	}
	if !esResp.Found {
		return nil, errors.New(string(response.Body))
	}

	return esResp, nil
}

// Delete used to delete document by id
func (c *ElasticsearchClient) Delete(indexName, id string) (*DeleteResponse, error) {
	if c.Config.IndexPrefix != "" {
		indexName = c.Config.IndexPrefix + indexName
	}
	url := c.Config.Endpoint + "/" + indexName + "/doc/" + id

	log.Debug("delete doc: ", url)

	req := util.NewDeleteRequest(url)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	response, err := util.ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	log.Trace("delete response: ", string(response.Body))

	esResp := &DeleteResponse{}
	err = json.Unmarshal(response.Body, esResp)
	if err != nil {
		return &DeleteResponse{}, err
	}
	if esResp.Result != "deleted" {
		return nil, errors.New(string(response.Body))
	}

	return esResp, nil
}

// Count used to count how many docs in one index
func (c *ElasticsearchClient) Count(indexName string) (*CountResponse, error) {

	if c.Config.IndexPrefix != "" {
		indexName = c.Config.IndexPrefix + indexName
	}

	url := c.Config.Endpoint + "/" + indexName + "/_count"

	log.Debug("doc count: ", url)

	req := util.NewGetRequest(url)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	response, err := util.ExecuteRequest(req)

	if err != nil {
		return nil, err
	}

	log.Trace("count response: ", string(response.Body))

	esResp := &CountResponse{}
	err = json.Unmarshal(response.Body, esResp)
	if err != nil {
		return &CountResponse{}, err
	}

	return esResp, nil
}

// Search used to execute a search query
func (c *ElasticsearchClient) Search(indexName string, query *SearchRequest) (*SearchResponse, error) {

	if query.From < 0 {
		query.From = 0
	}
	if query.Size <= 0 {
		query.Size = 10
	}

	js, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	return c.SearchWithRawQueryDSL(indexName, js)
}

func (c *ElasticsearchClient) SearchWithRawQueryDSL(indexName string, queryDSL []byte) (*SearchResponse, error) {

	if c.Config.IndexPrefix != "" {
		indexName = c.Config.IndexPrefix + indexName
	}

	url := c.Config.Endpoint + "/" + indexName + "/_search"

	log.Debug("search: ", url)

	req := util.NewPostRequest(url, queryDSL)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	response, err := util.ExecuteRequest(req)
	if err != nil {
		return nil, err
	}

	log.Trace("search response: ", string(queryDSL), ",", string(response.Body))
	log.Info("search response: ", string(queryDSL), ",", string(response.Body))

	esResp := &SearchResponse{}
	err = json.Unmarshal(response.Body, esResp)
	if err != nil {
		return &SearchResponse{}, err
	}

	return esResp, nil
}
