package elastic

import (
	"fmt"
	"strings"

	"github.com/xirtah/gopa-framework/core/errors"
	"github.com/xirtah/gopa-framework/core/index"
	api "github.com/xirtah/gopa-framework/core/persist"
	"github.com/xirtah/gopa-framework/core/util"
)

type ElasticORM struct {
	Client *index.ElasticsearchClient
}

func getIndex(any interface{}) string {
	return util.GetTypeName(any, true)
}

func getID(any interface{}) string {
	return util.GetFieldValueByTagName(any, "index", "id")
}

func (handler ElasticORM) Get(o interface{}) error {

	response, err := handler.Client.Get(getIndex(o), getID(o))
	if err != nil {
		return err
	}

	//TODO improve performance
	str := util.ToJson(response.Source, false)
	return util.FromJson(str, o)
}

func (handler ElasticORM) GetBy(field string, value interface{}, t interface{}, to interface{}) (error, api.Result) {

	query := api.Query{}
	query.Size = 1 //only return the first result -- TODO: Confirm if it is okay to only return 1 result - LZRBEAR
	query.Conds = api.And(api.Eq(field, value))
	return handler.Search(t, to, &query)
}

func (handler ElasticORM) Save(o interface{}) error {
	_, err := handler.Client.Index(getIndex(o), getID(o), o)
	return err
}

func (handler ElasticORM) Update(o interface{}) error {
	return handler.Save(o)
}

func (handler ElasticORM) Delete(o interface{}) error {
	_, err := handler.Client.Delete(getIndex(o), getID(o))
	return err
}

func (handler ElasticORM) Count(o interface{}) (int, error) {
	countResponse, err := handler.Client.Count(getIndex(o))
	if err != nil {
		return 0, err
	}
	return countResponse.Count, err
}

func getQuery(c1 *api.Cond) interface{} {
	switch c1.QueryType {
	case api.Match:
		q := index.TermQuery{}
		q.SetTerm(c1.Field, c1.Value)
		return q
	case api.RangeGt:
		q := index.RangeQuery{}
		q.Gt(c1.Field, c1.Value)
		return q
	case api.RangeGte:
		q := index.RangeQuery{}
		q.Gte(c1.Field, c1.Value)
		return q
	case api.RangeLt:
		q := index.RangeQuery{}
		q.Lt(c1.Field, c1.Value)
		return q
	case api.RangeLte:
		q := index.RangeQuery{}
		q.Lte(c1.Field, c1.Value)
		return q
	}
	panic(errors.Errorf("invalid query: %s", c1))
}

func (handler ElasticORM) Search(t interface{}, to interface{}, q *api.Query) (error, api.Result) {

	var err error

	request := index.SearchRequest{}

	request.From = q.From
	request.Size = q.Size

	if q.Conds != nil && len(q.Conds) > 0 {
		request.Query = &index.Query{}
		boolQuery := index.BoolQuery{}

		for _, c1 := range q.Conds {
			switch c1.BoolType {
			case api.Must:
				//boolQuery.Filter = append(boolQuery.Filter, getQuery(c1))
				//TODO: Clean up commented out code
				boolQuery.Must = append(boolQuery.Must, getQuery(c1))
				break
			case api.MustNot:
				boolQuery.MustNot = append(boolQuery.MustNot, getQuery(c1))
				break
			case api.Should:
				//boolQuery.Should = append(boolQuery.Should, q)
				break
			case api.Filter:
				boolQuery.Filter = append(boolQuery.Filter, getQuery(c1))
				break
			}

			request.Query.Bool = &boolQuery
		}
	}

	result := api.Result{}
	searchResponse, err := handler.Client.Search(getIndex(t), &request)
	if err != nil {
		return err, result
	}

	array := []interface{}{}

	for _, doc := range searchResponse.Hits.Hits {
		array = append(array, doc.Source)
	}

	result.Result = array
	result.Total = searchResponse.Hits.Total

	return err, result
}

//TODO: Clean up this function, it is quite hacky at the moment
func (handler ElasticORM) GroupBy(o interface{}, selectField, groupField string, haveQuery string, haveValue interface{}) (error, map[string]interface{}) {

	request := index.SearchRequest{}

	request.Aggs = &index.Aggs{}
	request.Size = 0
	aggs := index.Aggs{}
	q := index.TermsQuery{}
	q.SetTerm("field", selectField)

	//HACK
	if haveValue != nil {
		s := strings.Split(haveQuery, " ")
		request.Query = &index.Query{}
		request.Query.Bool = &index.BoolQuery{}
		q1 := index.TermQuery{}
		q1.SetTerm(s[0], haveValue)
		request.Query.Bool.Filter = append(request.Query.Bool.Filter, q1)
	}
	//END HACK

	aggs.Terms = &q
	request.Aggs = &aggs

	searchResponse, err := handler.Client.Search(getIndex(o), &request)

	result := map[string]interface{}{}

	for _, doc := range searchResponse.Aggregations {
		for _, dock := range doc.Buckets {

			var slc string
			switch v := dock.Key.(type) {
			case string:
				slc = v
			case fmt.Stringer:
				slc = v.String()
			default:
				slc = fmt.Sprintf("%v", v)
			}

			result[slc] = dock.DocCount

		}
	}

	return err, result
}
