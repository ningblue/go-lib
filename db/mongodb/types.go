package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
)

// Errors defines
var (
	ErrTransactionStated   = errors.New("transaction already started")
	ErrTransactionNotFound = errors.New("not in transaction environment")
	ErrDocumentNotFound    = errors.New("document not found")
	ErrNoDocumentsInResult = errors.New("mongo: no documents in result")
	ErrDuplicated          = errors.New("duplicated")
	ErrSessionNotStarted   = errors.New("session is not started")

	UpdateOpAddToSet = "addToSet"
	UpdateOpPull     = "pull"
)

type FindOpts struct {
	WithObjectID *bool
	WithCount    *bool
	NoDeleteAt   *bool
}

// NewFindOpts TODO
func NewFindOpts() *FindOpts {
	defaultTrue := true
	return &FindOpts{
		WithObjectID: &defaultTrue,
		NoDeleteAt:   &defaultTrue,
	}
}

// SetWithObjectID TODO
func (f *FindOpts) SetWithObjectID(bl bool) *FindOpts {
	f.WithObjectID = &bl
	return f
}

// SetWithCount TODO
func (f *FindOpts) SetWithCount(bl bool) *FindOpts {
	f.WithCount = &bl
	return f
}

// AggregateOpts TODO
type AggregateOpts struct {
	AllowDiskUse *bool
}

// NewAggregateOpts TODO
func NewAggregateOpts() *AggregateOpts {
	return &AggregateOpts{}
}

// SetAllowDiskUse TODO
func (a *AggregateOpts) SetAllowDiskUse(bl bool) *AggregateOpts {
	a.AllowDiskUse = &bl
	return a
}

// ModeUpdate  根据不同的操作符去更新数据
type ModeUpdate struct {
	Op  string
	Doc interface{}
}

// Index define the DB index struct
type Index struct {
	Keys                    bson.D                 `json:"keys" bson:"key"`
	Name                    string                 `json:"name" bson:"name"`
	Unique                  bool                   `json:"unique" bson:"unique"`
	Background              bool                   `json:"background" bson:"background"`
	ExpireAfterSeconds      int32                  `json:"expire_after_seconds" bson:"expire_after_seconds,omitempty"`
	PartialFilterExpression map[string]interface{} `json:"partialFilterExpression" bson:"partialFilterExpression"`
}

// Filter condition alias name
type Filter interface{}

// Table TODO
type Table interface {
	// Find 查询多个并反序列化到 Result
	Find(filter Filter, opts ...*FindOpts) FindInterface
	// AggregateOne 聚合查询
	AggregateOne(pipeline interface{}, result interface{}) error
	AggregateAll(pipeline interface{}, result interface{}, opts ...*AggregateOpts) error
	// Insert 插入数据, docs 可以为 单个数据 或者 多个数据
	Insert(docs interface{}) error
	// Update 更新数据
	Update(filter Filter, doc interface{}) error
	// Upsert TODO
	// update or insert data
	Upsert(filter Filter, doc interface{}) error
	// UpdateMultiModel  data based on operators.
	UpdateMultiModel(filter Filter, updateModel ...ModeUpdate) error

	// Delete 默认软删除数据
	Delete(filter Filter) error
	// HardDelete 硬删除
	HardDelete(filter Filter) error

	// CreateIndex 创建索引
	CreateIndex(index Index) error
	// BatchCreateIndexes 批量创建索引
	BatchCreateIndexes(index []Index) error

	// DropIndex 移除索引
	DropIndex(indexName string) error
	// Indexes 查询索引
	Indexes(ctx context.Context) ([]Index, error)

	// AddColumn 添加字段
	AddColumn(column string, value interface{}) error
	// RenameColumn 重命名字段
	RenameColumn(filter Filter, oldName, newColumn string) error
	// DropColumn 移除字段
	DropColumn(field string) error
	// DropColumns 根据条件移除字段
	DropColumns(filter Filter, fields []string) error

	// DropDocsColumn remove a column by the name for doc use filter
	DropDocsColumn(field string, filter Filter) error

	// Distinct Finds the distinct values for a specified field across a single collection or view and returns the results in an
	// field the field for which to return distinct values.
	// filter query that specifies the documents from which to retrieve the distinct values.
	Distinct(field string, filter Filter) ([]interface{}, error)

	// DeleteMany delete document, return number of documents that were deleted.
	DeleteMany(filter Filter) (uint64, error)
	// HardDeleteMany hard delete document
	HardDeleteMany(filter Filter) (uint64, error)

	// UpdateMany update document, return number of documents that were modified.
	UpdateMany(filter Filter, doc interface{}) (uint64, error)
}

// FindInterface operation interface
type FindInterface interface {
	// Fields 设置查询字段
	Fields(fields ...string) FindInterface
	// Sort 设置查询排序
	Sort(sort string) FindInterface
	// Start 设置限制查询上标
	Start(start int64) FindInterface
	// Limit 设置查询数量
	Limit(limit int64) FindInterface
	// All 查询多个
	All(result interface{}) error
	// One 查询单个
	One(result interface{}) error
	// Count 统计数量(非事务)
	Count() (uint64, error)
	// List 查询多个, start 等于0的时候，返回满足条件的行数
	List(result interface{}) (int64, error)

	Option(opts ...*FindOpts)
}
