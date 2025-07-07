package mongodb

import (
	"context"
	"github.com/ningblue/go-lib/convert"

	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	collName string // 集合名
	*MongoDriver
}

func (c *Collection) AggregateOne(pipeline interface{}, result interface{}) error {
	opt := getCollectionOption(c.ctx)
	cursor, err := c.client.Database(c.dbName).Collection(c.collName, opt).Aggregate(c.ctx, pipeline)
	if err != nil {
		return err

	}
	defer cursor.Close(c.ctx)
	for cursor.Next(c.ctx) {
		return cursor.Decode(result)
	}
	return ErrDocumentNotFound

}

// AggregateAll 聚合查询所有数据
// DOTO  暂时没有支持软删除
func (c *Collection) AggregateAll(pipeline interface{}, result interface{}, opts ...*AggregateOpts) error {
	var aggregateOption *options.AggregateOptions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if opt.AllowDiskUse != nil {
			aggregateOption = &options.AggregateOptions{AllowDiskUse: opt.AllowDiskUse}
		}
	}
	opt := getCollectionOption(c.ctx)

	cursor, err := c.client.Database(c.dbName).Collection(c.collName, opt).Aggregate(c.ctx, pipeline, aggregateOption)
	if err != nil {
		return err
	}
	defer cursor.Close(c.ctx)
	return cursor.All(c.ctx, result)
}

func (c *Collection) Insert(docs interface{}) error {
	// 加锁
	c.mu.Lock()
	defer c.mu.Unlock()

	rows := convert.ConvertToInterfaceSlice(docs)
	// 插入数据
	_, err := c.client.Database(c.dbName).Collection(c.collName).InsertMany(c.ctx, rows)
	return err
}

func (c *Collection) Update(filter Filter, doc interface{}) error {
	if filter == nil {
		filter = bson.M{}
	}
	data := bson.M{"$set": doc}
	_, err := c.client.Database(c.dbName).Collection(c.collName).UpdateMany(c.ctx, filter, data)
	return err
}

// Upsert 数据存在更新数据，否则新加数据。
// 注意：该接口非原子操作，可能存在插入多条相同数据的风险。
func (c *Collection) Upsert(filter Filter, doc interface{}) error {
	// set upsert option
	doUpsert := true
	replaceOpt := &options.UpdateOptions{
		Upsert: &doUpsert,
	}
	data := bson.M{"$set": doc}
	_, err := c.client.Database(c.dbName).Collection(c.collName).UpdateOne(c.ctx, filter, data, replaceOpt)
	return err

}

func (c *Collection) UpdateMultiModel(filter Filter, updateModel ...ModeUpdate) error {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) Delete(filter Filter) error {
	_, err := c.DeleteMany(filter)
	return err
}
func (c *Collection) HardDelete(filter Filter) error {
	_, err := c.HardDeleteMany(filter)
	return err
}
func (c *Collection) CreateIndex(index Index) error {
	createIndexInfo, err := buildIndex(index)
	if err != nil {
		return err
	}
	indexView := c.client.Database(c.dbName).Collection(c.collName).Indexes()
	_, err = indexView.CreateOne(c.ctx, createIndexInfo)
	if err != nil {
		if strings.Contains(err.Error(), "all indexes already exist") ||
			strings.Contains(err.Error(), "already exists with a different name") {
			return nil
		}

	}
	return err
}
func buildIndex(index Index) (mongo.IndexModel, error) {
	createIndexOpt := &options.IndexOptions{
		Unique:                  &index.Unique,
		PartialFilterExpression: index.PartialFilterExpression,
	}
	if index.Name != "" {
		createIndexOpt.Name = &index.Name
	}

	if index.ExpireAfterSeconds != 0 {
		createIndexOpt.SetExpireAfterSeconds(index.ExpireAfterSeconds)
	}

	keys := index.Keys
	for idx, key := range keys {
		val, err := GetInt32ByInterface(key.Value)
		if err != nil {
			return mongo.IndexModel{}, err
		}
		key.Value = val
		keys[idx] = key
	}

	return mongo.IndexModel{
		Keys:    keys,
		Options: createIndexOpt,
	}, nil
}
func GetInt32ByInterface(a interface{}) (int32, error) {
	id := int32(0)
	var err error
	switch val := a.(type) {
	case int:
		id = int32(val)
	case int32:
		id = val
	case int64:
		id = int32(val)
	case json.Number:
		var tmpID int64
		tmpID, err = val.Int64()
		id = int32(tmpID)
	case float64:
		id = int32(val)
	case float32:
		id = int32(val)
	default:
		err = errors.New("not numeric")

	}
	return id, err
}
func (c *Collection) BatchCreateIndexes(index []Index) error {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) DropIndex(indexName string) error {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) Indexes(ctx context.Context) ([]Index, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) AddColumn(column string, value interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) RenameColumn(filter Filter, oldName, newColumn string) error {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) DropColumn(field string) error {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) DropColumns(filter Filter, fields []string) error {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) DropDocsColumn(field string, filter Filter) error {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) Distinct(field string, filter Filter) ([]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Collection) hasDeletedAtField() (bool, error) {
	filter := bson.M{"deleted_at": bson.M{"$exists": true}}
	opts := options.Find().SetLimit(1) // 只获取一个文档进行检查

	cursor, err := c.client.Database(c.dbName).Collection(c.collName).Find(c.ctx, filter, opts)
	if err != nil {
		return false, err
	}
	defer cursor.Close(c.ctx)
	exist := cursor.Next(c.ctx)
	return exist, nil

}

// softDeleteMany
func (c *Collection) softDeleteMany(filter Filter) (uint64, error) {
	// 通过读取集合的 deleted_at字段,如果有这个字段就可以进行软删除
	// 也就是 将 deleted_at 字段设置为当前时间
	if filter == nil {
		filter = bson.M{}
	}
	data := bson.M{"$set": bson.M{"deleted_at": time.Now()}}
	result, err := c.client.Database(c.dbName).Collection(c.collName).UpdateMany(c.ctx, filter, data)
	if err != nil {
		return 0, err
	}
	return uint64(result.ModifiedCount), nil
}
func (c *Collection) HardDeleteMany(filter Filter) (uint64, error) {
	return c.deleteMany(filter)
}
func (c *Collection) deleteMany(filter Filter) (uint64, error) {
	var deleteCount uint64
	result, err := c.client.Database(c.dbName).Collection(c.collName).DeleteMany(c.ctx, filter)
	if err != nil {
		return 0, err
	}
	deleteCount = uint64(result.DeletedCount)
	return deleteCount, nil
}

// DeleteMany 删除多个文档
// 通过软删除的方式删除文档,如果需要硬删除,请使用 HardDeleteMany
func (c *Collection) DeleteMany(filter Filter) (uint64, error) {
	// 默认调用软删除
	return c.softDeleteMany(filter)
}

func (c *Collection) UpdateMany(filter Filter, doc interface{}) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

// Find 查询多个并反序列化到 Result
func (c *Collection) Find(filter Filter, opts ...*FindOpts) FindInterface {
	find := &Find{
		Collection: c,
		filter:     filter,
		option:     *NewFindOpts(),
	}
	find.Option(opts...)
	if find.option.NoDeleteAt != nil && *find.option.NoDeleteAt {
		find.AddDeletedAtFilter()
	}
	return find
}

// addDeletedAtFilterToMap adds the deleted_at filter to the map
func addDeletedAtFilterToMap(filter map[string]interface{}) map[string]interface{} {
	if filter == nil {
		filter = map[string]interface{}{"deleted_at": nil}
	} else {
		//  如果已经有了 $or并且也有了 $and, 则将 筛选条件加入到 $and 中
		if orConditions, ok := filter["$or"].([]interface{}); ok {
			if andConditions, ok := filter["$and"].([]interface{}); ok {
				filter["$and"] = append(andConditions, map[string]interface{}{"deleted_at": nil})
			} else {
				// 如果没有and条件 就需要把之前的or条件加到and条件中,并增加筛选
				filter["$and"] = []interface{}{
					map[string]interface{}{"$or": orConditions},
					map[string]interface{}{"deleted_at": nil},
				}
				// 需要删除之前的or筛选
				delete(filter, "$or")
			}
			// 如果没有or条件,直接加入到and条件中
		} else if andConditions, ok := filter["$and"].([]interface{}); ok {
			filter["$and"] = append(andConditions, map[string]interface{}{"deleted_at": nil})
		} else {
			// 都没有筛选就直接添加条件
			filter["deleted_at"] = nil
		}
	}
	return filter
}

// AddDeletedAtFilter adds the deleted_at filter to the query
func (f *Find) AddDeletedAtFilter() {
	if f.filter == nil {
		// 默认 查询所有 deleted_at is null
		f.filter = bson.M{"deleted_at": nil}
	} else {
		switch filter := f.filter.(type) {
		case MapStr:
			f.filter = MapStr(addDeletedAtFilterToMap(filter))
		case map[string]interface{}:
			f.filter = addDeletedAtFilterToMap(filter)
		case bson.M:
			f.filter = bson.M(addDeletedAtFilterToMap(filter))
		}
	}
}

type Find struct {
	*Collection
	projection map[string]int // 查询字段
	filter     Filter         // 查询条件
	start      int64          // 查询上标
	limit      int64          // 查询数量
	sort       bson.D         // 查询排序
	option     FindOpts       // 查询选项
}

func (f *Find) generateMongoOption() *options.FindOptions {
	findOpts := &options.FindOptions{}
	if f.projection == nil {
		f.projection = make(map[string]int, 0)
	}
	if f.option.WithObjectID != nil && *f.option.WithObjectID {
		// mongodb 要求，当有字段设置未1, 不设置都不显示
		// 没有设置projection 的时候，返回所有字段
		if len(f.projection) > 0 {
			f.projection["_id"] = 1
		}
	} else {
		if _, exists := f.projection["_id"]; !exists {
			f.projection["_id"] = 0
		}
	}
	if len(f.projection) != 0 {
		findOpts.Projection = f.projection
	}

	if f.start != 0 {
		findOpts.SetSkip(f.start)
	}
	if f.limit != 0 {
		findOpts.SetLimit(f.limit)
	}
	if len(f.sort) != 0 {
		findOpts.SetSort(f.sort)
	}

	return findOpts
}

// Fields 设置查询字段
func (f *Find) Fields(fields ...string) FindInterface {
	for _, field := range fields {
		if len(field) <= 0 {
			continue
		}
		f.projection[field] = 1
	}
	return f
}

// Start 设置限制查询上标
func (f *Find) Start(start int64) FindInterface {
	f.start = start
	return f
}

// Sort 设置查询排序
func (f *Find) Sort(sort string) FindInterface {
	if sort != "" {
		sortArr := strings.Split(sort, ",")
		f.sort = bson.D{}
		for _, sortItem := range sortArr {
			sortItemArr := strings.Split(strings.TrimSpace(sortItem), ":")
			sortKey := strings.TrimLeft(sortItemArr[0], "+-")
			if len(sortItemArr) == 2 {
				sortDescFlag := strings.TrimSpace(sortItemArr[1])
				if sortDescFlag == "-1" {
					f.sort = append(f.sort, bson.E{Key: sortKey, Value: -1})
				} else {
					f.sort = append(f.sort, bson.E{Key: sortKey, Value: 1})
				}
			} else {
				if strings.HasPrefix(sortItemArr[0], "-") {
					f.sort = append(f.sort, bson.E{Key: sortKey, Value: -1})
				} else {
					f.sort = append(f.sort, bson.E{Key: sortKey, Value: 1})
				}
			}
		}
	}

	return f
}

// Limit 设置查询数量
func (f *Find) Limit(limit int64) FindInterface {
	f.limit = limit
	return f
}

func (f *Find) All(result interface{}) error {
	findOpts := f.generateMongoOption()
	// 查询条件为空时候，mongodb 不返回数据
	if f.filter == nil {
		f.filter = bson.M{}
	}
	cursor, err := f.client.Database(f.dbName).Collection(f.collName).Find(f.ctx, f.filter, findOpts)
	if err != nil {
		return err
	}
	return cursor.All(f.ctx, result)
}

func (f *Find) One(result interface{}) error {
	// TODO 后续支持软删除查询
	err := f.client.Database(f.dbName).Collection(f.collName).FindOne(f.ctx, f.filter).Decode(result)
	//Bug  这里 使用 Find 通过游标 的时候 获取结果的时候,_id 字段会丢失,所以暂时使用 findOne 方法
	//cursor, err := f.client.Database(f.dbName).Collection(f.collName).Find(f.ctx, f.filter, findOpts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrDocumentNotFound
		}
		return err
	}
	return nil
}

func (f *Find) Count() (uint64, error) {
	if f.filter == nil {
		f.filter = bson.M{}
	}
	cnt, err := f.client.Database(f.dbName).Collection(f.collName).CountDocuments(f.ctx, f.filter)
	if err != nil {
		return 0, err
	}
	return uint64(cnt), nil

}

func (f *Find) List(result interface{}) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Find) Option(opts ...*FindOpts) {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if opt.WithObjectID != nil {
			f.option.WithObjectID = opt.WithObjectID
		}
		if opt.WithCount != nil {
			f.option.WithCount = opt.WithCount
		}
		if opt.NoDeleteAt != nil {
			f.option.NoDeleteAt = opt.NoDeleteAt
		}
	}
}

const (
	// reference doc:
	// https://docs.mongodb.com/manual/core/read-preference-staleness/#replica-set-read-preference-max-staleness
	// this is the minimum value of maxStalenessSeconds allowed.
	// specifying a smaller maxStalenessSeconds value will raise an error. Clients estimate secondaries’ staleness
	// by periodically checking the latest write date of each replica set member. Since these checks are infrequent,
	// the staleness estimate is coarse. Thus, clients cannot enforce a maxStalenessSeconds value of less than
	// 90 seconds.
	maxStalenessSeconds = 90 * time.Second
)

func decodeCursorIntoSlice(ctx context.Context, cursor *mongo.Cursor, result interface{}) error {
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		return errors.New("result argument must be a slice address")
	}

	elemt := resultv.Elem().Type().Elem()
	slice := reflect.MakeSlice(resultv.Elem().Type(), 0, 10)
	for cursor.Next(ctx) {
		elemp := reflect.New(elemt)
		if err := cursor.Decode(elemp.Interface()); nil != err {
			return err
		}
		slice = reflect.Append(slice, elemp.Elem())
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	resultv.Elem().Set(slice)
	return nil
}
