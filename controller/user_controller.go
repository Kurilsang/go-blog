package controller

import (
	"go_test/dto"
	"go_test/global"
	"go_test/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var userService = service.NewUserService()

// GetMyProfile 获取自己的用户资料
func GetMyProfile(ctx *gin.Context) {
	// 从JWT中间件获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID格式错误"})
		return
	}

	profile, err := userService.GetProfile(uid)
	if err != nil {
		if err.Error() == "用户不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "获取用户资料成功",
		"data":    profile,
	})
}

// UpdateMyProfile 更新自己的用户资料
func UpdateMyProfile(ctx *gin.Context) {
	// 从JWT中间件获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID格式错误"})
		return
	}

	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	err := userService.UpdateProfile(uid, req)
	if err != nil {
		if err.Error() == "邮箱已被其他用户使用" || err.Error() == "没有需要更新的字段" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新用户资料成功"})
}

// ChangeMyPassword 修改自己的密码
func ChangeMyPassword(ctx *gin.Context) {
	// 从JWT中间件获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID格式错误"})
		return
	}

	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	err := userService.ChangePassword(uid, req)
	if err != nil {
		if err.Error() == "旧密码错误" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if err.Error() == "用户不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// GetUserProfile 获取指定用户资料（管理员或用户本人）
func GetUserProfile(ctx *gin.Context) {
	// 获取目标用户ID
	targetUserIDStr := ctx.Param("id")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式错误"})
		return
	}

	// 从JWT中间件获取当前用户信息
	currentUserID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	currentUID, ok := currentUserID.(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID格式错误"})
		return
	}

	// 获取用户角色（更新key名称）
	userRole, exists := ctx.Get("userRole")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户角色信息"})
		return
	}

	role, ok := userRole.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户角色格式错误"})
		return
	}

	isAdmin := (role == global.RoleAdmin)

	profile, err := userService.GetUserByID(uint(targetUserID), currentUID, isAdmin)
	if err != nil {
		if err.Error() == "权限不足，只能查看自己的资料" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else if err.Error() == "用户不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "获取用户资料成功",
		"data":    profile,
	})
}

// UpdateUserProfile 更新指定用户资料（管理员功能）
func UpdateUserProfile(ctx *gin.Context) {
	// 获取目标用户ID
	targetUserIDStr := ctx.Param("id")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式错误"})
		return
	}

	var req dto.AdminUpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	err = userService.UpdateUserByAdmin(uint(targetUserID), req)
	if err != nil {
		if err.Error() == "邮箱已被其他用户使用" || err.Error() == "没有需要更新的字段" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新用户资料成功"})
}

// GetAllUsers 获取所有用户列表（管理员功能）
func GetAllUsers(ctx *gin.Context) {
	users, err := userService.GetAllUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "获取用户列表成功",
		"data":    users,
		"total":   len(users),
	})
}
