package controllers

import (
	"home-monitor-backend/models"
	"home-monitor-backend/services"
	"home-monitor-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) *UserController {
	return &UserController{userService: userService}
}

// UserRegister godoc
// @Summary Register new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.UserRegisterRequest true "User register request"
// @Success 201 {object} models.UserRegisterResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /user/register [post]
func (ctrl *UserController) UserRegister(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input models.UserRegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors})
		return
	}

	user, statusCode, err := ctrl.userService.UserRegister(input, userUUID.(uuid.UUID))
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, models.UserRegisterResponse{
		UUID:      user.UUID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// UserLogin godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.UserLoginRequest true "User login request"
// @Success 200 {object} models.UserLoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /user/login [post]
func (ctrl *UserController) UserLogin(c *gin.Context) {
	var input models.UserLoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	user, token, statusCode, err := ctrl.userService.UserLogin(input)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.UserLoginResponse{
		UUID:     user.UUID,
		Username: user.Username,
		Token:    "Bearer " + token,
	})
}

// UserProfile godoc
// @Summary Get user profile
// @Description Retrieve the profile of the authenticated user
// @Tags users
// @Produce json
// @Success 200 {object} models.UserProfileResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /user/profile [get]
func (ctrl *UserController) UserProfile(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, statusCode, err := ctrl.userService.UserProfile(userUUID.(uuid.UUID))
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, models.UserProfileResponse{
		UUID:      user.UUID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// UserUpdate godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.UserUpdateRequest true "User update request"
// @Success 200 {object} models.UserRegisterResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /user/update [put]
func (ctrl *UserController) UserUpdate(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input models.UserUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors})
		return
	}

	user, statusCode, err := ctrl.userService.UserUpdate(userUUID.(uuid.UUID), &input)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, models.UserRegisterResponse{
		UUID:      user.UUID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}
