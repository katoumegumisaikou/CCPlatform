package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/pkg/util"
	"fmt"
)

// DictService 数据字典业务逻辑层。
type DictService struct {
	repo *repository.DictRepo
}

func NewDictService() *DictService {
	return &DictService{repo: repository.NewDictRepo()}
}

// --- Dict ---

func (s *DictService) CreateDict(d *model.Dict) error {
	d.DictID = util.GenerateID()
	return s.repo.CreateDict(d)
}

func (s *DictService) GetDictByID(id string) (*model.Dict, error) {
	return s.repo.GetDictByID(id)
}

func (s *DictService) GetDictByType(dictType string) (*model.Dict, error) {
	return s.repo.GetDictByType(dictType)
}

func (s *DictService) UpdateDict(d *model.Dict) error {
	_, err := s.repo.GetDictByID(d.DictID)
	if err != nil {
		return fmt.Errorf("dict not found: %w", err)
	}
	return s.repo.UpdateDict(d)
}

func (s *DictService) DeleteDict(id string) error {
	dict, err := s.repo.GetDictByID(id)
	if err != nil {
		return fmt.Errorf("dict not found: %w", err)
	}
	// Delete all items under this dict type
	items, _ := s.repo.GetItemsByType(dict.DictType)
	for _, item := range items {
		s.repo.DeleteItem(item.ItemID)
	}
	return s.repo.DeleteDict(id)
}

func (s *DictService) ListDict(page, size int) ([]model.Dict, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.ListDict(page, size)
}

// --- DictItem ---

func (s *DictService) CreateItem(item *model.DictItem) error {
	item.ItemID = util.GenerateID()
	return s.repo.CreateItem(item)
}

func (s *DictService) GetItemByID(id string) (*model.DictItem, error) {
	return s.repo.GetItemByID(id)
}

func (s *DictService) GetItemsByType(dictType string) ([]model.DictItem, error) {
	return s.repo.GetItemsByType(dictType)
}

func (s *DictService) UpdateItem(item *model.DictItem) error {
	_, err := s.repo.GetItemByID(item.ItemID)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}
	return s.repo.UpdateItem(item)
}

func (s *DictService) DeleteItem(id string) error {
	return s.repo.DeleteItem(id)
}
