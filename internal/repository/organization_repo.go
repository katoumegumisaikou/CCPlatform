package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// OrganizationRepo 组织架构数据访问层。
type OrganizationRepo struct {
	db *gorm.DB
}

func NewOrganizationRepo() *OrganizationRepo {
	return &OrganizationRepo{db: DB}
}

func (r *OrganizationRepo) Create(org *model.Organization) error {
	return r.db.Create(org).Error
}

func (r *OrganizationRepo) GetByID(id string) (*model.Organization, error) {
	var org model.Organization
	err := r.db.Where("org_id = ?", id).First(&org).Error
	return &org, err
}

func (r *OrganizationRepo) Update(org *model.Organization) error {
	return r.db.Save(org).Error
}

func (r *OrganizationRepo) Delete(id string) error {
	return r.db.Where("org_id = ?", id).Delete(&model.Organization{}).Error
}

// List 分页查询组织列表，支持按类型和关键词筛选。
func (r *OrganizationRepo) List(page, size int, orgType, keyword string) ([]model.Organization, int64, error) {
	var orgs []model.Organization
	var total int64
	query := r.db.Model(&model.Organization{})
	if orgType != "" {
		query = query.Where("org_type = ?", orgType)
	}
	if keyword != "" {
		query = query.Where("org_name LIKE ? OR org_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Offset((page - 1) * size).Limit(size).Order("sort_order ASC, create_time DESC").Find(&orgs).Error
	return orgs, total, err
}

// GetByParent 查询指定父节点下的子组织。
func (r *OrganizationRepo) GetByParent(parentID string) ([]model.Organization, error) {
	var orgs []model.Organization
	err := r.db.Where("parent_id = ?", parentID).Order("sort_order ASC").Find(&orgs).Error
	return orgs, err
}

// GetAll 查询所有启用状态的组织（用于构建树）。
func (r *OrganizationRepo) GetAll() ([]model.Organization, error) {
	var orgs []model.Organization
	err := r.db.Where("status = 1").Order("sort_order ASC, create_time ASC").Find(&orgs).Error
	return orgs, err
}

// HasChildren 检查指定组织是否有子节点。
func (r *OrganizationRepo) HasChildren(orgID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Organization{}).Where("parent_id = ?", orgID).Count(&count).Error
	return count > 0, err
}
