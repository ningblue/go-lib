package gorm

import (
	"context"
	"fmt"
	"reflect"

	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CURDHooks[T any] interface {
	BeforeCreate(ctx context.Context, entity *T) error
	AfterCreate(ctx context.Context, entity *T) error
	BeforeUpdate(ctx context.Context, id uint, updateData map[string]interface{}) error
	AfterUpdate(ctx context.Context, id uint, updateData map[string]interface{}) error
	BeforeDelete(ctx context.Context, id uint) error
	AfterDelete(ctx context.Context, id uint) error
}

type DefaultHooks[T any] struct{}

func (h *DefaultHooks[T]) BeforeCreate(ctx context.Context, entity *T) error {
	return nil
}

func (h *DefaultHooks[T]) AfterCreate(ctx context.Context, entity *T) error {
	return nil
}

func (h *DefaultHooks[T]) BeforeUpdate(ctx context.Context, id uint, updateData map[string]interface{}) error {
	return nil
}

func (h *DefaultHooks[T]) AfterUpdate(ctx context.Context, id uint, updateData map[string]interface{}) error {
	return nil
}

func (h *DefaultHooks[T]) BeforeDelete(ctx context.Context, id uint) error {
	return nil
}

func (h *DefaultHooks[T]) AfterDelete(ctx context.Context, id uint) error {
	return nil
}

type QueryParams struct {
	Query       map[string]interface{} `json:"query" form:"query"`               // 查询条件
	JsonQuery   map[string]interface{} `json:"json_query" form:"json_query"`     // json类型查询条件
	Page        int                    `json:"page" form:"page"`                 // 当前页码
	PageSize    int                    `json:"page_size" form:"page_size"`       // 每页显示数量
	Sort        string                 `json:"sort" form:"sort"`                 // 排序字段
	SearchValue map[string]interface{} `json:"search_value" form:"search_value"` // 模糊搜索
}

func NewQueryParams(page, pageSize int, sort string, query map[string]interface{}, searchValue map[string]interface{}) QueryParams {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if sort == "" {
		sort = "created_at desc"
	}
	if query == nil {
		query = make(map[string]interface{})
	}

	if searchValue == nil {
		searchValue = make(map[string]interface{})
	}
	return QueryParams{
		Query:       query,
		JsonQuery:   make(map[string]interface{}),
		Page:        page,
		PageSize:    pageSize,
		Sort:        sort,
		SearchValue: searchValue,
	}
}

type CURDRepository[T any] interface {
	CreateByTx(tx *gorm.DB, entity *T) error
	Create(ctx context.Context, entity *T) error
	UpdateByTx(tx *gorm.DB, id uint, updateData map[string]interface{}) error
	List(ctx context.Context, query QueryParams, preloads ...string) ([]*T, int64, error)
	ListReturnDB(ctx context.Context, query QueryParams) (*gorm.DB, error)
	GetByID(ctx context.Context, id uint, preload ...string) (*T, error)
	GetByUUID(ctx context.Context, uuid string) (*T, error)
	GetBy(ctx context.Context, field string, value interface{}) (*T, error)
	GetByNumberID(ctx context.Context, id string, preload ...string) (*T, error)
	Update(ctx context.Context, id uint, updateData map[string]interface{}) error
	UpdateIns(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
	All(ctx context.Context) ([]*T, error)
	AllByFields(ctx context.Context, fields map[string]interface{}, preload ...string) ([]*T, error)
}

type GenericCURDRepo[T any] struct {
	DB    *gorm.DB
	hooks CURDHooks[T]
}

var _ CURDRepository[any] = (*GenericCURDRepo[any])(nil)

func NewGenericCURDRepo[T any](db *gorm.DB, hooks CURDHooks[T]) *GenericCURDRepo[T] {
	if hooks == nil {
		hooks = &DefaultHooks[T]{}
	}
	return &GenericCURDRepo[T]{
		DB:    db,
		hooks: hooks,
	}
}

func (r *GenericCURDRepo[T]) Create(ctx context.Context, entity *T) error {
	if err := r.hooks.BeforeCreate(ctx, entity); err != nil {
		return err
	}
	if err := r.DB.WithContext(ctx).Create(entity).Error; err != nil {
		return err
	}
	if err := r.hooks.AfterCreate(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (r *GenericCURDRepo[T]) CreateByTx(tx *gorm.DB, entity *T) error {
	if err := tx.Create(entity).Error; err != nil {
		return err
	}
	return nil
}

func (r *GenericCURDRepo[T]) UpdateByTx(tx *gorm.DB, id uint, updateData map[string]interface{}) error {
	if err := r.hooks.BeforeUpdate(context.Background(), id, updateData); err != nil {
		return err
	}
	if err := tx.Model(new(T)).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}
	if err := r.hooks.AfterUpdate(context.Background(), id, updateData); err != nil {
		return err
	}
	return nil
}

func (r *GenericCURDRepo[T]) ListReturnDB(ctx context.Context, query QueryParams) (*gorm.DB, error) {
	db := r.DB.WithContext(ctx)
	if query.Sort != "" {
		db = db.Order(query.Sort)
	}
	if query.SearchValue != nil && len(query.SearchValue) > 0 {
		var searchConditions []string
		var searchValues []interface{}
		for key, value := range query.SearchValue {
			searchConditions = append(searchConditions, fmt.Sprintf("%s LIKE ?", key))
			searchValues = append(searchValues, fmt.Sprintf("%%%v%%", value))
		}
		searchQuery := strings.Join(searchConditions, " OR ")
		db = db.Where(searchQuery, searchValues...)
	}
	if err := db.Model(new(T)).Where(query.Query).Error; err != nil {
		hlog.Error("list data failed ", err)
		return nil, err
	}
	return db, nil
}

func (r *GenericCURDRepo[T]) List(ctx context.Context, query QueryParams, preloads ...string) ([]*T, int64, error) {
	var entities []*T
	var total int64
	db := r.DB.WithContext(ctx)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.Sort != "" {
		db = db.Order(query.Sort)
	}
	// 遍历 JSON 查询条件并构建查询
	for field, value := range query.JsonQuery {
		if reflect.TypeOf(value).Kind() == reflect.Slice {
			db = db.Where(datatypes.JSONArrayQuery(field).Contains(value))
		} else if reflect.TypeOf(value).Kind() == reflect.Map {
			// 处理嵌套的JSON结构
			if innerStrMap, ok := value.(map[string]string); ok {
				// 向后兼容原有的string-string映射
				for innerKey, innerValue := range innerStrMap {
					db = db.Where(datatypes.JSONQuery(field).Equals(innerValue, innerKey))
				}
			}
		} else if value == nil {
			// 查询字段为null的情况
			db = db.Where(fmt.Sprintf("JSON_EXTRACT(%s, '$') IS NULL", field))
		} else {
			// 处理原始情况 - 非JSON查询
			db = db.Where(fmt.Sprintf("%s = ?", field), value)
		}
	}
	if query.SearchValue != nil && len(query.SearchValue) > 0 {
		for key, value := range query.SearchValue {
			db = db.Or(fmt.Sprintf("%s LIKE ? ", key), fmt.Sprintf("%%%v%%", value))
		}
	}
	if err := db.Model(new(T)).Where(query.Query).Count(&total).Error; err != nil {
		hlog.Error("list data failed ", err)
		return nil, 0, err
	}
	for _, field := range preloads {
		db = db.Preload(field)
	}
	if err := db.Model(new(T)).Where(query.Query).Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&entities).Error; err != nil {
		hlog.Error("list data failed ", err)
		return nil, 0, err
	}
	return entities, total, nil
}

func (r *GenericCURDRepo[T]) All(ctx context.Context) ([]*T, error) {
	var entities []*T
	if err := r.DB.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GenericCURDRepo[T]) GetByID(ctx context.Context, id uint, preload ...string) (*T, error) {
	var entity T
	db := r.DB.WithContext(ctx)

	for _, field := range preload {
		db = db.Preload(field)
	}

	if err := db.First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetBy .
func (r *GenericCURDRepo[T]) GetBy(ctx context.Context, field string, value interface{}) (*T, error) {
	var entity T
	if err := r.DB.WithContext(ctx).Where(fmt.Sprintf("%s = ?", field), value).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetByFields .
func (r *GenericCURDRepo[T]) GetByFields(ctx context.Context, fields map[string]interface{}, preload ...string) (*T, error) {
	var entity T
	db := r.DB.WithContext(ctx)
	for _, field := range preload {
		db = db.Preload(field)
	}
	if err := r.DB.WithContext(ctx).Where(fields).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GenericCURDRepo[T]) GetByNumberID(ctx context.Context, id string, preload ...string) (*T, error) {
	var entity T
	db := r.DB.WithContext(ctx)

	for _, field := range preload {
		db = db.Preload(field)
	}

	if err := db.Where("number_id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GenericCURDRepo[T]) GetByUUID(ctx context.Context, uuid string) (*T, error) {
	var entity T
	if err := r.DB.WithContext(ctx).Where("uuid = ?", uuid).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GenericCURDRepo[T]) Update(ctx context.Context, id uint, updateData map[string]interface{}) error {
	if err := r.hooks.BeforeUpdate(ctx, id, updateData); err != nil {
		return err
	}

	if err := r.DB.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(updateData).Error; err != nil {
		return err
	}

	if err := r.hooks.AfterUpdate(ctx, id, updateData); err != nil {
		return err
	}
	return nil
}

// UpdateIns
func (r *GenericCURDRepo[T]) UpdateIns(ctx context.Context, entity *T) error {
	if err := r.hooks.BeforeUpdate(ctx, 0, nil); err != nil {
		return err
	}
	if err := r.DB.WithContext(ctx).Save(entity).Error; err != nil {
		return err
	}

	if err := r.hooks.AfterUpdate(ctx, 0, nil); err != nil {
		return err
	}
	return nil
}

func (r *GenericCURDRepo[T]) Delete(ctx context.Context, id uint) error {
	if err := r.hooks.BeforeDelete(ctx, id); err != nil {
		return err
	}

	if err := r.DB.WithContext(ctx).Delete(new(T), id).Error; err != nil {
		return err
	}

	if err := r.hooks.AfterDelete(ctx, id); err != nil {
		return err
	}
	return nil
}

// AllByFields
func (r *GenericCURDRepo[T]) AllByFields(ctx context.Context, fields map[string]interface{}, preload ...string) ([]*T, error) {
	var entity []*T
	db := r.DB.WithContext(ctx)
	for _, field := range preload {
		db = db.Preload(field)
	}
	if err := r.DB.WithContext(ctx).Where(fields).Find(&entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

// 递归处理嵌套的JSON结构
func processNestedJSON(db *gorm.DB, field string, parentKey string, value interface{}) *gorm.DB {
	if nestedMap, ok := value.(map[string]interface{}); ok {
		for nestedKey, nestedValue := range nestedMap {
			if reflect.TypeOf(nestedValue).Kind() == reflect.Map {
				// 继续递归处理
				processNestedJSON(db, field, parentKey+"."+nestedKey, nestedValue)
			} else {
				// 到达叶节点，构建查询
				db = db.Where(datatypes.JSONQuery(field).Equals(nestedValue, parentKey, nestedKey))
			}
		}
	}
	return db
}
