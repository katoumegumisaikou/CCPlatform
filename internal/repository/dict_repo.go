package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// DictRepo 数据字典数据访问层。
type DictRepo struct {
	db *gorm.DB
}

func NewDictRepo() *DictRepo {
	return &DictRepo{db: DB}
}

// --- Dict ---

func (r *DictRepo) CreateDict(d *model.Dict) error {
	return r.db.Create(d).Error
}

func (r *DictRepo) GetDictByID(id string) (*model.Dict, error) {
	var d model.Dict
	err := r.db.Where("dict_id = ?", id).First(&d).Error
	return &d, err
}

func (r *DictRepo) GetDictByType(dictType string) (*model.Dict, error) {
	var d model.Dict
	err := r.db.Where("dict_type = ?", dictType).First(&d).Error
	return &d, err
}

func (r *DictRepo) UpdateDict(d *model.Dict) error {
	return r.db.Save(d).Error
}

func (r *DictRepo) DeleteDict(id string) error {
	return r.db.Where("dict_id = ?", id).Delete(&model.Dict{}).Error
}

func (r *DictRepo) ListDict(page, size int) ([]model.Dict, int64, error) {
	var dicts []model.Dict
	var total int64
	r.db.Model(&model.Dict{}).Count(&total)
	err := r.db.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&dicts).Error
	return dicts, total, err
}

// --- DictItem ---

func (r *DictRepo) CreateItem(item *model.DictItem) error {
	return r.db.Create(item).Error
}

func (r *DictRepo) GetItemByID(id string) (*model.DictItem, error) {
	var item model.DictItem
	err := r.db.Where("item_id = ?", id).First(&item).Error
	return &item, err
}

func (r *DictRepo) GetItemsByType(dictType string) ([]model.DictItem, error) {
	var items []model.DictItem
	err := r.db.Where("dict_type = ?", dictType).Order("sort_order ASC").Find(&items).Error
	return items, err
}

func (r *DictRepo) UpdateItem(item *model.DictItem) error {
	return r.db.Save(item).Error
}

func (r *DictRepo) DeleteItem(id string) error {
	return r.db.Where("item_id = ?", id).Delete(&model.DictItem{}).Error
}
