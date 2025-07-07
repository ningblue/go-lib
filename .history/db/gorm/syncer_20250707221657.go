package gorm

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	libUtils "github.com/ningblue/go-lib/utils"
)

type ResourceSyncer[T any] struct {
	DB            *gorm.DB
	GetNewData    func(ctx context.Context) (map[string]*T, error)
	GetOldData    func() (map[string]*T, error)
	UpdateQuery   func(oldItem *T, newItem *T) *gorm.DB
	CompareFields []string
	DeleteQuery   func(id string) *gorm.DB
}

// Sync 同步资源
func (s *ResourceSyncer[T]) Sync(ctx context.Context) error {
	// Step 1: 获取新数据
	newData, err := s.GetNewData(ctx)
	if err != nil {
		return err
	}
	// Step 2: 获取旧数据
	oldData, err := s.GetOldData()
	if err != nil {
		return err
	}
	// Step 3: 对比数据并应用变更
	var InsertData []*T
	for id, newItem := range newData {
		if oldItem, exists := oldData[id]; !exists {
			InsertData = append(InsertData, newItem)
		} else {
			// 对比并更新
			isSame, err := libUtils.CompareStruct(newItem, oldItem, s.CompareFields)
			if err != nil {
				return err
			}
			if !isSame {
				fmt.Printf("Syncer: update item id: %s, newItem: %v, oldItem: %v\n", id, newItem, oldItem)
				if err := s.UpdateQuery(oldItem, newItem).Updates(newItem).Error; err != nil {
					return err
				}
			}
		}
	}

	if len(InsertData) > 0 {
		// 批量插入新数据
		if err := s.DB.CreateInBatches(InsertData, 1000).Error; err != nil {
			return err
		}
	}
	// Step 4: 删除不存在于新数据中的旧数据
	for id, oldItem := range oldData {
		if _, exists := newData[id]; !exists {
			if err := s.DeleteQuery(id).Delete(oldItem).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
