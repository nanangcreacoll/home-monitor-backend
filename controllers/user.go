package controllers

import (
	"home-monitor-backend/models"
	"home-monitor-backend/services"
	"home-monitor-backend/utils"
	"net/http"
	"strconv"

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
// @Router /user [post]
func (ctrl *UserController) UserRegister(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var input models.UserRegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	user, statusCode, err := ctrl.userService.UserRegister(c, input, userUUID.(uuid.UUID))
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
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

	user, token, statusCode, err := ctrl.userService.UserLogin(c, input)
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
// @Router /user [get]
func (ctrl *UserController) UserProfile(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	user, statusCode, err := ctrl.userService.UserProfile(c, userUUID.(uuid.UUID))
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
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
// @Router /user [put]
func (ctrl *UserController) UserUpdate(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var input models.UserUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	user, statusCode, err := ctrl.userService.UserUpdate(c, userUUID.(uuid.UUID), &input)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.UserRegisterResponse{
		UUID:      user.UUID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// UserDelete godoc
// @Summary Delete user account
// @Description Delete the account of the authenticated user
// @Tags users
// @Produce json
// @Success 200 {object} models.ResponseMessage
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /user [delete]
func (ctrl *UserController) UserDelete(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var input models.UserDeleteRequest
	if err := c.ShouldBindJSON(&input); err != nil && err.Error() != "EOF" {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	deleteUUID := uuid.Nil
	if input.UUID != uuid.Nil {
		deleteUUID = input.UUID
	}

	user, statusCode, err := ctrl.userService.UserDelete(c, userUUID.(uuid.UUID), deleteUUID)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.ResponseMessage{Message: "User " + user.Username + " deleted successfully"})
}

// UserList godoc
// @Summary List users
// @Description Retrieve a list of users (admin only)
// @Tags users
// @Produce json
// @Success 200 {object} models.UserListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /users [get]
func (ctrl *UserController) UserList(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	lengthStr := c.Query("length")
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		length = 0
	}

	users, statusCode, err := ctrl.userService.UserList(c, userUUID.(uuid.UUID), length)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	var userProfiles []models.UserProfileResponse
	for _, user := range users {
		userProfiles = append(userProfiles, models.UserProfileResponse{
			UUID:      user.UUID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	c.JSON(statusCode, models.UserListResponse{Users: userProfiles})
}
