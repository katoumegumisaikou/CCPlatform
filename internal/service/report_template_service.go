package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"fmt"
)

// ReportTemplateService 报表模板业务逻辑层。
type ReportTemplateService struct {
	repo *repository.ReportTemplateRepo
}

func NewReportTemplateService() *ReportTemplateService {
	return &ReportTemplateService{repo: repository.NewReportTemplateRepo()}
}

func (s *ReportTemplateService) Create(tpl *model.ReportTemplate) error {
	tpl.TplID = generateID()
	return s.repo.Create(tpl)
}

func (s *ReportTemplateService) GetByID(id string) (*model.ReportTemplate, error) {
	return s.repo.GetByID(id)
}

func (s *ReportTemplateService) Update(tpl *model.ReportTemplate) error {
	_, err := s.repo.GetByID(tpl.TplID)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}
	return s.repo.Update(tpl)
}

func (s *ReportTemplateService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *ReportTemplateService) List(page, size int) ([]model.ReportTemplate, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size)
}
