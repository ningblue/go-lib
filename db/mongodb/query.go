package mongodb

import (
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

const (
	// PageName TODO
	PageName = "page"
	// PageSort TODO
	PageSort = "sort"
	// PageStart TODO
	PageStart = "start"
	// PageLimit TODO
	PageLimit = "limit"
	// DBFields TODO
	DBFields = "fields"
	// DBQueryCondition TODO
	DBQueryCondition = "condition"
)

// BasePage for paging query
type BasePage struct {
	Sort        string `json:"sort,omitempty"`
	Limit       int64  `json:"limit,omitempty"`
	Start       int64  `json:"start"`
	EnableCount bool   `json:"enable_count,omitempty"`
}

// SearchModuleCondition search module condition
type SearchModuleCondition struct {
	Condition map[string]interface{} `json:"condition"`
	Page      BasePage               `json:"page"`
	Fields    []string               `json:"fields,omitempty"`
	//SetID     int64                  `json:"bk_set_id"`
}

// QueryCondition the common query condition definition
type QueryCondition struct {
	Fields    []string `json:"fields"`
	Page      BasePage `json:"page"`
	Condition MapStr   `json:"condition"`
	// 非必填，只能用来查时间，且与Condition是与关系
	//TimeCondition  *TimeCondition `json:"time_condition,omitempty"`
	DisableCounter bool `json:"disable_counter"`
}

func NewQueryCondition(pageNum, pageSize int64, sort string, filter map[string]interface{}) QueryCondition {
	if filter == nil {
		filter = make(map[string]interface{})
	}
	return QueryCondition{
		Fields: make([]string, 0),
		Page: BasePage{
			Sort:        sort,
			Limit:       pageSize,
			Start:       (pageNum - 1) * pageSize,
			EnableCount: false,
		},
		Condition:      filter,
		DisableCounter: false,
	}
}

// AggregateConditionNew 构建aggregate查询条件
type AggregateConditionNew struct {
	Sort      string                   `json:"sort"`
	Skip      int64                    `json:"skip"`
	Limit     int64                    `json:"limit"`
	Condition []map[string]interface{} `json:"condition"`
}

// AggregateCondition 构建aggregate查询条件
type AggregateCondition struct {
	Match       bson.M `json:"match"`
	Sort        string `json:"sort"`
	Skip        int64  `json:"skip"`
	Limit       int64  `json:"limit"`
	Condition   bson.M `json:"condition"`
	Lookup      bson.M `json:"lookup"`
	AddFields   bson.M `json:"addFields"`
	GraphLookup bson.M `json:"graphLookup"`
}

func NewAggregateConditionNew(pageNum, pageSize int64, sort string, condition []map[string]interface{}) AggregateConditionNew {
	return AggregateConditionNew{
		Sort:      sort,
		Skip:      (pageNum - 1) * pageSize,
		Limit:     pageSize,
		Condition: condition,
	}
}
func NewAggregateCondition(match bson.M, sort string, pageNum, pageSize int64, lookup, graphLookup bson.M, addFields bson.M) AggregateCondition {
	return AggregateCondition{
		Match:       match,
		Sort:        sort,
		Skip:        (pageNum - 1) * pageSize,
		Limit:       pageSize,
		Lookup:      lookup,
		GraphLookup: graphLookup,
		AddFields:   addFields,
	}
}

type MapStr map[string]interface{}

// QueryResult common query result
type QueryResult struct {
	Count uint64   `json:"count"`
	Info  []MapStr `json:"info"`
}

// BuildFilter 构建过滤器函数
func BuildFilter(condition string, fields []string) bson.M {
	var conditions []bson.M

	for _, field := range fields {
		fieldValue := strings.Split(field, ":")
		if len(fieldValue) == 2 {
			conditions = append(conditions, bson.M{fieldValue[0]: bson.M{"$regex": fieldValue[1], "$options": "i"}})
		}
	}

	filter := bson.M{}
	if condition == "or" {
		filter["$or"] = conditions
	} else {
		filter["$and"] = conditions
	}
	return filter
}
