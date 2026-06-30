package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// NotifyTemplateRepo 通知模板数据访问层。
type NotifyTemplateRepo struct {
	db *gorm.DB
}

func NewNotifyTemplateRepo() *NotifyTemplateRepo {
	return &NotifyTemplateRepo{db: DB}
}

func (r *NotifyTemplateRepo) Create(tpl *model.NotifyTemplate) error {
	return r.db.Create(tpl).Error
}

func (r *NotifyTemplateRepo) GetByID(id string) (*model.NotifyTemplate, error) {
	var tpl model.NotifyTemplate
	err := r.db.Where("tpl_id = ?", id).First(&tpl).Error
	return &tpl, err
}

func (r *NotifyTemplateRepo) Update(tpl *model.NotifyTemplate) error {
	return r.db.Save(tpl).Error
}

func (r *NotifyTemplateRepo) Delete(id string) error {
	return r.db.Where("tpl_id = ?", id).Delete(&model.NotifyTemplate{}).Error
}

func (r *NotifyTemplateRepo) List(page, size int, tplType string) ([]model.NotifyTemplate, int64, error) {
	var tpls []model.NotifyTemplate
	var total int64
	query := r.db.Model(&model.NotifyTemplate{})
	if tplType != "" {
		query = query.Where("tpl_type = ?", tplType)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&tpls).Error
	return tpls, total, err
}

// GetActive 查询所有启用的通知模板。
func (r *NotifyTemplateRepo) GetActive() ([]model.NotifyTemplate, error) {
	var tpls []model.NotifyTemplate
	err := r.db.Where("status = 1").Find(&tpls).Error
	return tpls, err
}
