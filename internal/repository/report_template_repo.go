package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// ReportTemplateRepo 报表模板数据访问层。
type ReportTemplateRepo struct {
	db *gorm.DB
}

func NewReportTemplateRepo() *ReportTemplateRepo {
	return &ReportTemplateRepo{db: DB}
}

func (r *ReportTemplateRepo) Create(tpl *model.ReportTemplate) error {
	return r.db.Create(tpl).Error
}

func (r *ReportTemplateRepo) GetByID(id string) (*model.ReportTemplate, error) {
	var tpl model.ReportTemplate
	err := r.db.Where("tpl_id = ?", id).First(&tpl).Error
	return &tpl, err
}

func (r *ReportTemplateRepo) Update(tpl *model.ReportTemplate) error {
	return r.db.Save(tpl).Error
}

func (r *ReportTemplateRepo) Delete(id string) error {
	return r.db.Where("tpl_id = ?", id).Delete(&model.ReportTemplate{}).Error
}

func (r *ReportTemplateRepo) List(page, size int) ([]model.ReportTemplate, int64, error) {
	var tpls []model.ReportTemplate
	var total int64
	query := r.db.Model(&model.ReportTemplate{})
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("create_time DESC").Find(&tpls).Error
	return tpls, total, err
}
