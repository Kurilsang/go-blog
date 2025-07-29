package service

import (
	"fmt"
	"go_test/dto"
	"go_test/global"
	"go_test/model"
	"go_test/utils"

	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// GetProfile 获取用户资料业务逻辑
func (s *UserService) GetProfile(userID uint) (*dto.UserProfileVO, error) {
	var user model.User
	if err := global.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}

	return &dto.UserProfileVO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Nickname: user.Nickname,
		Bio:      user.Bio,
		Phone:    user.Phone,
		Role:     user.Role,
		Status:   user.Status,
		Created:  user.CreatedAt.Format("2006-01-02 15:04:05"),
		Updated:  user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateProfile 更新用户资料业务逻辑
func (s *UserService) UpdateProfile(userID uint, req dto.UpdateProfileRequest) error {
	// 构建更新数据
	updateData := map[string]interface{}{}

	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		var existingUser model.User
		if err := global.DB.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			return fmt.Errorf("邮箱已被其他用户使用")
		}
		updateData["email"] = req.Email
	}

	if req.Avatar != "" {
		updateData["avatar"] = req.Avatar
	}

	if req.Nickname != "" {
		updateData["nickname"] = req.Nickname
	}

	if req.Bio != "" {
		updateData["bio"] = req.Bio
	}

	if req.Phone != "" {
		updateData["phone"] = req.Phone
	}

	// 如果没有需要更新的字段
	if len(updateData) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}

	// 执行更新
	if err := global.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updateData).Error; err != nil {
		return fmt.Errorf("更新用户资料失败: %v", err)
	}

	return nil
}

// ChangePassword 修改密码业务逻辑
func (s *UserService) ChangePassword(userID uint, req dto.ChangePasswordRequest) error {
	// 获取用户当前密码
	var user model.User
	if err := global.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("用户不存在")
		}
		return err
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, user.Password) {
		return fmt.Errorf("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败")
	}

	// 更新密码
	if err := global.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("密码更新失败: %v", err)
	}

	return nil
}

// GetUserByID 根据用户ID获取用户资料（用于管理员查看）
func (s *UserService) GetUserByID(targetUserID uint, currentUserID uint, isAdmin bool) (*dto.UserProfileVO, error) {
	// 权限检查：只有管理员或本人可以查看
	if !isAdmin && targetUserID != currentUserID {
		return nil, fmt.Errorf("权限不足，只能查看自己的资料")
	}

	return s.GetProfile(targetUserID)
}

// UpdateUserByAdmin 管理员更新用户资料业务逻辑
func (s *UserService) UpdateUserByAdmin(targetUserID uint, req dto.AdminUpdateUserRequest) error {
	// 构建更新数据
	updateData := map[string]interface{}{}

	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		var existingUser model.User
		if err := global.DB.Where("email = ? AND id != ?", req.Email, targetUserID).First(&existingUser).Error; err == nil {
			return fmt.Errorf("邮箱已被其他用户使用")
		}
		updateData["email"] = req.Email
	}

	if req.Avatar != "" {
		updateData["avatar"] = req.Avatar
	}

	if req.Nickname != "" {
		updateData["nickname"] = req.Nickname
	}

	if req.Bio != "" {
		updateData["bio"] = req.Bio
	}

	if req.Phone != "" {
		updateData["phone"] = req.Phone
	}

	// 管理员可以修改角色和状态
	if req.Role != "" {
		updateData["role"] = req.Role
	}

	if req.Status != "" {
		updateData["status"] = req.Status
	}

	// 如果没有需要更新的字段
	if len(updateData) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}

	// 执行更新
	if err := global.DB.Model(&model.User{}).Where("id = ?", targetUserID).Updates(updateData).Error; err != nil {
		return fmt.Errorf("更新用户资料失败: %v", err)
	}

	return nil
}

// GetAllUsers 获取所有用户列表（管理员功能）
func (s *UserService) GetAllUsers() ([]dto.UserProfileVO, error) {
	var users []model.User
	if err := global.DB.Find(&users).Error; err != nil {
		return nil, err
	}

	vos := make([]dto.UserProfileVO, 0, len(users))
	for _, user := range users {
		vos = append(vos, dto.UserProfileVO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Nickname: user.Nickname,
			Bio:      user.Bio,
			Phone:    user.Phone,
			Role:     user.Role,
			Status:   user.Status,
			Created:  user.CreatedAt.Format("2006-01-02 15:04:05"),
			Updated:  user.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return vos, nil
}
