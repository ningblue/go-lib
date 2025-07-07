package mongodb

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type indexField struct {
	field    bson.E
	priority int
}

type indexInfo struct {
	fields     []indexField
	unique     bool
	primaryKey bool
}

func parseBsonAttachTag(tag string) (indexName string, unique bool, primaryKey bool, priority int) {
	parts := strings.Split(tag, ",")
	priority = 10 // default priority
	for _, part := range parts {
		if part == "primaryKey" {
			primaryKey = true
		} else if part == "unique" {
			unique = true
		}
		if strings.Contains(part, ":") {
			kv := strings.Split(part, ":")
			switch kv[0] {
			case "index":
				if len(kv) > 1 {
					indexName = kv[1]
				}
			case "priority":
				if len(kv) > 1 {
					priority = int(kv[1][0] - '0')
				}
			}
		}
	}
	return
}

// CreateIndexes 创建索引
//
//		举个例子:
//	 支持	设置bson_attach 字段 ,支持索引 优先级,唯一索引 复合索引
//		type User struct {
//			ID       primitive.ObjectID `bson:"_id" json:"id"`
//			UserName string             `bson:"name" json:"name" bson_attach:"index:index_user,priority:1,unique"`
//			Age      int                `bson:"age" json:"age"`
//			Email    string             `bson:"email" json:"email" bson_attach:"index:index_user,priority:2" `
//		}
func (m *MongoDriver) CreateIndexes(models []map[string]interface{}, delSameIndex bool) {
	for _, model := range models {
		m.createIndexForModel(model, delSameIndex)
	}
}

func (m *MongoDriver) createIndexForModel(modelMap map[string]interface{}, delSameIndex bool) {
	if modelMap == nil {
		return
	}
	if _, ok := modelMap["tableName"]; !ok {
		hlog.Error("tableName is required")
		return
	}
	if _, ok := modelMap["model"]; !ok {
		hlog.Error("model is required")
		return
	}
	collectionName := modelMap["tableName"].(string)
	model := modelMap["model"]
	collection := m.client.Database(m.dbName).Collection(collectionName)
	indexMap := make(map[string]*indexInfo)

	modelType := reflect.TypeOf(model)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		bsonField := field.Tag.Get("bson")
		if tag, ok := field.Tag.Lookup("bson_attach"); ok {
			indexName, unique, primaryKey, priority := parseBsonAttachTag(tag)
			if indexName != "" {
				if idx, exists := indexMap[indexName]; exists {
					idx.fields = append(idx.fields, indexField{
						field:    bson.E{Key: bsonField, Value: 1},
						priority: priority,
					})
					idx.unique = idx.unique || unique
					idx.primaryKey = idx.primaryKey || primaryKey
				} else {
					indexMap[indexName] = &indexInfo{
						fields: []indexField{
							{
								field:    bson.E{Key: bsonField, Value: 1},
								priority: priority,
							},
						},
						unique:     unique,
						primaryKey: primaryKey,
					}
				}
			}
		}
	}
	// 默认不删除,相同就报错
	if delSameIndex {
		// 获取现有索引列表
		existingIndexes, err := collection.Indexes().List(context.TODO())
		if err != nil {
			hlog.Error(fmt.Sprintf("Failed to list indexes for collection %s: %v", collectionName, err))
			return
		}

		var existingIndexNames []string
		for existingIndexes.Next(context.TODO()) {
			var index bson.M
			if err := existingIndexes.Decode(&index); err != nil {
				hlog.Error(fmt.Sprintf("Failed to decode index info for collection %s: %v", collectionName, err))
				return
			}
			if name, ok := index["name"].(string); ok {
				existingIndexNames = append(existingIndexNames, name)
			}
		}

		// 删除旧索引
		for indexName := range indexMap {
			for _, existingIndexName := range existingIndexNames {
				if existingIndexName == indexName {
					if _, err := collection.Indexes().DropOne(context.TODO(), existingIndexName); err != nil {
						hlog.Error(fmt.Sprintf("Failed to drop index %s for collection %s: %v", existingIndexName, collectionName, err))
						return
					}
					hlog.Infof(fmt.Sprintf("Dropped existing index %s for collection %s", existingIndexName, collectionName))
					break
				}
			}
		}
	}

	var indexModels []mongo.IndexModel
	for indexName, idx := range indexMap {
		newIndexName := indexName
		// 对字段按优先级排序
		sort.SliceStable(idx.fields, func(i, j int) bool {
			return idx.fields[i].priority < idx.fields[j].priority
		})
		var keys bson.D
		for _, f := range idx.fields {
			keys = append(keys, f.field)
		}
		indexModel := mongo.IndexModel{
			Keys: keys,
			Options: &options.IndexOptions{
				Name: &newIndexName,
			},
		}
		if idx.unique {
			indexModel.Options.SetUnique(true)
		}
		if idx.primaryKey {
			indexModel.Options.SetUnique(true)
		}
		indexModels = append(indexModels, indexModel)
	}

	if len(indexModels) > 0 {
		_, err := collection.Indexes().CreateMany(context.TODO(), indexModels)
		if err != nil {
			hlog.Error(fmt.Sprintf("Failed to create indexes for collection %s: %v", collectionName, err))
			return
		}
		hlog.Infof(fmt.Sprintf("Created new indexes for collection %s", collectionName))
	}
}
