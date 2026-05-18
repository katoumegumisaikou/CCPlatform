package service

import (
	"ccplatform/internal/model"
	"ccplatform/internal/repository"
	"fmt"
)

// OrganizationService 组织架构业务逻辑层。
type OrganizationService struct {
	repo *repository.OrganizationRepo
}

func NewOrganizationService() *OrganizationService {
	return &OrganizationService{repo: repository.NewOrganizationRepo()}
}

// OrgTreeNode 组织架构树节点。
type OrgTreeNode struct {
	*model.Organization
	Children []*OrgTreeNode `json:"children"`
}

func (s *OrganizationService) Create(org *model.Organization) error {
	org.OrgID = generateID()
	return s.repo.Create(org)
}

func (s *OrganizationService) GetByID(id string) (*model.Organization, error) {
	return s.repo.GetByID(id)
}

func (s *OrganizationService) Update(org *model.Organization) error {
	return s.repo.Update(org)
}

func (s *OrganizationService) Delete(id string) error {
	hasChildren, err := s.repo.HasChildren(id)
	if err != nil {
		return err
	}
	if hasChildren {
		return fmt.Errorf("organization has children, cannot delete")
	}
	return s.repo.Delete(id)
}

func (s *OrganizationService) List(page, size int, orgType, keyword string) ([]model.Organization, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return s.repo.List(page, size, orgType, keyword)
}

// GetTree 构建组织架构树。
func (s *OrganizationService) GetTree() ([]*OrgTreeNode, error) {
	orgs, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[string]*OrgTreeNode)
	var roots []*OrgTreeNode

	for i := range orgs {
		nodeMap[orgs[i].OrgID] = &OrgTreeNode{Organization: &orgs[i], Children: []*OrgTreeNode{}}
	}
	for _, node := range nodeMap {
		if node.ParentID == "" || nodeMap[node.ParentID] == nil {
			roots = append(roots, node)
		} else {
			parent := nodeMap[node.ParentID]
			parent.Children = append(parent.Children, node)
		}
	}
	return roots, nil
}
