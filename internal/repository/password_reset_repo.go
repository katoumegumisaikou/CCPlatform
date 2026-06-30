package repository

import (
	"ccplatform/internal/model"

	"gorm.io/gorm"
)

// PasswordResetRepo 密码重置数据访问层。
type PasswordResetRepo struct {
	db *gorm.DB
}

func NewPasswordResetRepo() *PasswordResetRepo {
	return &PasswordResetRepo{db: DB}
}

func (r *PasswordResetRepo) Create(pr *model.PasswordReset) error {
	return r.db.Create(pr).Error
}

func (r *PasswordResetRepo) GetByToken(token string) (*model.PasswordReset, error) {
	var pr model.PasswordReset
	err := r.db.Where("token = ?", token).First(&pr).Error
	return &pr, err
}

func (r *PasswordResetRepo) MarkUsed(token string) error {
	return r.db.Model(&model.PasswordReset{}).Where("token = ?", token).Update("used", 1).Error
}
