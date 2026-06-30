package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"ccplatform/pkg/util"
)

// NotifyService 通知模板业务逻辑层。
type NotifyService struct {
	repo *repository.NotifyTemplateRepo
}

func NewNotifyService() *NotifyService {
	return &NotifyService{repo: repository.NewNotifyTemplateRepo()}
}

func (s *NotifyService) Create(tpl *model.NotifyTemplate) error {
	tpl.TplID = util.GenerateID()
	return s.repo.Create(tpl)
}

func (s *NotifyService) Update(tpl *model.NotifyTemplate) error {
	return s.repo.Update(tpl)
}

func (s *NotifyService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *NotifyService) List(page, size int, tplType string) ([]model.NotifyTemplate, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, tplType)
}
