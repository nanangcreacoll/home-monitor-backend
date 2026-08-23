package tests

import (
	"context"
	"home-monitor-backend/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserRegister(t *testing.T) {
	ctx := context.Background()

	userRequest := models.UserRegisterRequest{
		Username: "testuser_" + uuid.New().String()[:8],
		Password: "testpassword",
	}

	adminCreate := &models.User{
		UUID:     uuid.New(),
		Username: "admin_" + uuid.New().String()[:8],
		Password: "adminpassword",
		Role:     models.UserRoleAdmin,
	}

	err := userRepository.UserCreate(ctx, adminCreate)
	assert.NoError(t, err, "Failed to create admin user")

	registeredUser, statusCode, err := userService.UserRegister(ctx, userRequest, adminCreate.UUID)
	assert.NoError(t, err, "User registration failed")
	assert.Equal(t, 201, statusCode, "Expected status code 201")

	assert.Equal(t, userRequest.Username, registeredUser.Username, "Expected username to match")

	_, statusCode, err = userService.UserRegister(ctx, userRequest, registeredUser.UUID)
	assert.Error(t, err, "Expected error when non-admin tries to register user")
	assert.Equal(t, 403, statusCode, "Expected status code 403 for non-admin registration attempt")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, registeredUser)
		userRepository.UserDelete(ctx, adminCreate)
	})
}

func TestUserLogin(t *testing.T) {
	ctx := context.Background()

	user := &models.User{
		UUID:     uuid.New(),
		Username: "loginuser_" + uuid.New().String()[:8],
		Password: "loginpassword",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, user)
	assert.NoError(t, err, "Failed to create user for login test")

	loginRequest := models.UserLoginRequest{
		Username: user.Username,
		Password: "loginpassword",
	}

	loggedInUser, token, statusCode, err := userService.UserLogin(ctx, loginRequest)
	assert.NoError(t, err, "User login failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")

	assert.Equal(t, user.Username, loggedInUser.Username, "Expected username to match")
	assert.NotEmpty(t, token, "Expected JWT token to be generated")

	invalidLoginRequest := models.UserLoginRequest{
		Username: user.Username,
		Password: "wrongpassword",
	}

	_, _, statusCode, err = userService.UserLogin(ctx, invalidLoginRequest)
	assert.Error(t, err, "Expected error for invalid login")
	assert.Equal(t, 401, statusCode, "Expected status code 401 for invalid login")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, user)
	})
}

func TestUserProfile(t *testing.T) {
	ctx := context.Background()

	user := &models.User{
		UUID:     uuid.New(),
		Username: "profileuser_" + uuid.New().String()[:8],
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, user)
	assert.NoError(t, err, "Failed to create user for profile test")

	profile, statusCode, err := userService.UserProfile(ctx, user.UUID)
	assert.NoError(t, err, "User profile retrieval failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, user.Username, profile.Username, "Expected username to match")
	assert.Equal(t, user.UUID, profile.UUID, "Expected UUID to match")

	nonExistentUUID := uuid.New()
	_, statusCode, err = userService.UserProfile(ctx, nonExistentUUID)
	assert.Error(t, err, "Expected error for non-existent user")
	assert.Equal(t, 401, statusCode, "Expected status code 401")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, user)
	})
}

func TestUserUpdate(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		UUID:     uuid.New(),
		Username: "updateadmin_" + uuid.New().String()[:8],
		Password: "adminpass123",
		Role:     models.UserRoleAdmin,
	}

	user := &models.User{
		UUID:     uuid.New(),
		Username: "updateuser_" + uuid.New().String()[:8],
		Password: "userpass123",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	err = userRepository.UserCreate(ctx, user)
	assert.NoError(t, err, "Failed to create regular user")

	newPassword := "newpassword123"
	updateRequest := &models.UserUpdateRequest{
		Username:    user.Username,
		Password:    "userpass123",
		NewPassword: newPassword,
	}

	updatedUser, statusCode, err := userService.UserUpdate(ctx, user.UUID, updateRequest)
	assert.NoError(t, err, "User password update failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")

	loginRequest := models.UserLoginRequest{
		Username: updatedUser.Username,
		Password: newPassword,
	}
	_, _, statusCode, err = userService.UserLogin(ctx, loginRequest)
	assert.NoError(t, err, "Login with new password failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")

	newUsername := "newusername_" + uuid.New().String()[:8]
	updateRequest = &models.UserUpdateRequest{
		Username:    updatedUser.Username,
		Password:    newPassword,
		NewUsername: newUsername,
	}

	updatedUser, statusCode, err = userService.UserUpdate(ctx, user.UUID, updateRequest)
	assert.NoError(t, err, "User username update failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, newUsername, updatedUser.Username, "Expected username to be updated")

	newRole := models.UserRoleAdmin
	updateRequest = &models.UserUpdateRequest{
		Username: updatedUser.Username,
		Password: newPassword,
		Role:     &newRole,
	}

	updatedUser, statusCode, err = userService.UserUpdate(ctx, admin.UUID, updateRequest)
	assert.NoError(t, err, "Admin role update failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, newRole, updatedUser.Role, "Expected role to be updated")

	anotherUser := &models.User{
		UUID:     uuid.New(),
		Username: "anotheruser_" + uuid.New().String()[:8],
		Password: "password123",
		Role:     models.UserRoleUser,
	}
	err = userRepository.UserCreate(ctx, anotherUser)
	assert.NoError(t, err, "Failed to create another user")

	updateRequest = &models.UserUpdateRequest{
		Username:    updatedUser.Username,
		Password:    newPassword,
		NewPassword: "hackpass123",
	}

	_, statusCode, err = userService.UserUpdate(ctx, anotherUser.UUID, updateRequest)
	assert.Error(t, err, "Expected error when non-admin tries to update another user")
	assert.Equal(t, 403, statusCode, "Expected status code 403")

	adminRoleUpdate := models.UserRoleUser
	updateRequest = &models.UserUpdateRequest{
		Username: admin.Username,
		Password: "adminpass123",
		Role:     &adminRoleUpdate,
	}

	_, statusCode, err = userService.UserUpdate(ctx, admin.UUID, updateRequest)
	assert.Error(t, err, "Expected error when admin tries to change own role")
	assert.Equal(t, 403, statusCode, "Expected status code 403")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, admin)
		userRepository.UserDelete(ctx, updatedUser)
		userRepository.UserDelete(ctx, anotherUser)
	})
}

func TestUserDelete(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		UUID:     uuid.New(),
		Username: "deleteadmin_" + uuid.New().String()[:8],
		Password: "adminpass123",
		Role:     models.UserRoleAdmin,
	}

	user := &models.User{
		UUID:     uuid.New(),
		Username: "deleteuser_" + uuid.New().String()[:8],
		Password: "userpass123",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	err = userRepository.UserCreate(ctx, user)
	assert.NoError(t, err, "Failed to create regular user")

	_, statusCode, err := userService.UserDelete(ctx, user.UUID, admin.UUID)
	assert.Error(t, err, "User other-deletion failed")
	assert.Equal(t, 403, statusCode, "Expected status code 403")

	deletedUser, statusCode, err := userService.UserDelete(ctx, user.UUID, uuid.Nil)
	assert.NoError(t, err, "User self-deletion failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, user.UUID, deletedUser.UUID, "Expected deleted user UUID to match")

	_, statusCode, err = userService.UserProfile(ctx, user.UUID)
	assert.Error(t, err, "Expected error for deleted user profile")
	assert.Equal(t, 401, statusCode, "Expected status code 401")

	deletedAdmin, statusCode, err := userService.UserDelete(ctx, admin.UUID, uuid.Nil)
	assert.NoError(t, err, "Admin self-deletion failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, admin.UUID, deletedAdmin.UUID, "Expected deleted admin UUID to match")

	_, statusCode, err = userService.UserProfile(ctx, admin.UUID)
	assert.Error(t, err, "Expected error for deleted admin profile")
	assert.Equal(t, 401, statusCode, "Expected status code 401")

	nonExistentUUID := uuid.New()
	_, statusCode, err = userService.UserDelete(ctx, nonExistentUUID, uuid.Nil)
	assert.Error(t, err, "Expected error for non-existent user")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestUserList(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		UUID:     uuid.New(),
		Username: "listadmin_" + uuid.New().String()[:8],
		Password: "adminpass123",
		Role:     models.UserRoleAdmin,
	}

	user := &models.User{
		UUID:     uuid.New(),
		Username: "listuser_" + uuid.New().String()[:8],
		Password: "userpass123",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	err = userRepository.UserCreate(ctx, user)
	assert.NoError(t, err, "Failed to create regular user")

	users, statusCode, err := userService.UserList(ctx, admin.UUID, 50)
	assert.NoError(t, err, "Admin user list failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.NotEmpty(t, users, "Expected user list to not be empty")

	_, statusCode, err = userService.UserList(ctx, user.UUID, 50)
	assert.Error(t, err, "Expected error when non-admin tries to list users")
	assert.Equal(t, 403, statusCode, "Expected status code 403")

	nonExistentUUID := uuid.New()
	_, statusCode, err = userService.UserList(ctx, nonExistentUUID, 50)
	assert.Error(t, err, "Expected error for non-existent user")
	assert.Equal(t, 401, statusCode, "Expected status code 401")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, admin)
		userRepository.UserDelete(ctx, user)
	})
}

func TestUserRegisterDuplicateUsername(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		UUID:     uuid.New(),
		Username: "dupeadmin_" + uuid.New().String()[:8],
		Password: "adminpass123",
		Role:     models.UserRoleAdmin,
	}

	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	username := "dupeuser_" + uuid.New().String()[:8]
	userRequest := models.UserRegisterRequest{
		Username: username,
		Password: "password123",
	}

	firstUser, statusCode, err := userService.UserRegister(ctx, userRequest, admin.UUID)
	assert.NoError(t, err, "First user registration failed")
	assert.Equal(t, 201, statusCode, "Expected status code 201")

	_, statusCode, err = userService.UserRegister(ctx, userRequest, admin.UUID)
	assert.Error(t, err, "Expected error for duplicate username")
	assert.Equal(t, 409, statusCode, "Expected status code 409")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, admin)
		userRepository.UserDelete(ctx, firstUser)
	})
}

func TestUserLoginNonExistentUser(t *testing.T) {
	ctx := context.Background()

	loginRequest := models.UserLoginRequest{
		Username: "nonexistent_" + uuid.New().String()[:8],
		Password: "password123",
	}

	_, _, statusCode, err := userService.UserLogin(ctx, loginRequest)
	assert.Error(t, err, "Expected error for non-existent user login")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestUserProfileNonExistent(t *testing.T) {
	ctx := context.Background()

	nonExistentUUID := uuid.New()
	_, statusCode, err := userService.UserProfile(ctx, nonExistentUUID)
	assert.Error(t, err, "Expected error for non-existent user profile")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestUserUpdateNonAdminCannotChangePassword(t *testing.T) {
	ctx := context.Background()

	user1 := &models.User{
		UUID:     uuid.New(),
		Username: "user1_" + uuid.New().String()[:8],
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	user2 := &models.User{
		UUID:     uuid.New(),
		Username: "user2_" + uuid.New().String()[:8],
		Password: "password456",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, user1)
	assert.NoError(t, err, "Failed to create user1")

	err = userRepository.UserCreate(ctx, user2)
	assert.NoError(t, err, "Failed to create user2")

	updateRequest := &models.UserUpdateRequest{
		Username:    user1.Username,
		Password:    "wrongpassword",
		NewPassword: "newpassword123",
	}

	_, statusCode, err := userService.UserUpdate(ctx, user2.UUID, updateRequest)
	assert.Error(t, err, "Expected error when non-owner tries to update user")
	assert.Equal(t, 403, statusCode, "Expected status code 403")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, user1)
		userRepository.UserDelete(ctx, user2)
	})
}

func TestUserUpdateNonExistentUser(t *testing.T) {
	ctx := context.Background()

	nonExistentUUID := uuid.New()
	updateRequest := &models.UserUpdateRequest{
		Username:    "someuser",
		Password:    "password123",
		NewPassword: "newpassword123",
	}

	_, statusCode, err := userService.UserUpdate(ctx, nonExistentUUID, updateRequest)
	assert.Error(t, err, "Expected error for non-existent user update")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestUserUpdateWrongPassword(t *testing.T) {
	ctx := context.Background()

	user := &models.User{
		UUID:     uuid.New(),
		Username: "wrongpass_" + uuid.New().String()[:8],
		Password: "correctpassword",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, user)
	assert.NoError(t, err, "Failed to create user")

	updateRequest := &models.UserUpdateRequest{
		Username:    user.Username,
		Password:    "wrongpassword",
		NewPassword: "newpassword123",
	}

	_, statusCode, err := userService.UserUpdate(ctx, user.UUID, updateRequest)
	assert.Error(t, err, "Expected error with wrong password")
	assert.Equal(t, 401, statusCode, "Expected status code 401")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, user)
	})
}

func TestUserUpdateDuplicateUsername(t *testing.T) {
	ctx := context.Background()

	user1 := &models.User{
		UUID:     uuid.New(),
		Username: "user1_unique_" + uuid.New().String()[:8],
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	admin := &models.User{
		UUID:     uuid.New(),
		Username: "admin_unique_" + uuid.New().String()[:8],
		Password: "adminpass456",
		Role:     models.UserRoleAdmin,
	}

	err := userRepository.UserCreate(ctx, user1)
	assert.NoError(t, err, "Failed to create user1")

	err = userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin")

	newUsername := "user1_newname_" + uuid.New().String()[:8]
	updateRequest := &models.UserUpdateRequest{
		Username:    user1.Username,
		Password:    "password123",
		NewUsername: newUsername,
	}

	updatedUser, statusCode, err := userService.UserUpdate(ctx, admin.UUID, updateRequest)
	assert.NoError(t, err, "Admin should be able to update user's username")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, newUsername, updatedUser.Username, "Expected username to be updated")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, admin)
		userRepository.UserDelete(ctx, updatedUser)
	})
}

func TestUserRegisterWithNonAdminUser(t *testing.T) {
	ctx := context.Background()

	normalUser := &models.User{
		UUID:     uuid.New(),
		Username: "normaluser_" + uuid.New().String()[:8],
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, normalUser)
	assert.NoError(t, err, "Failed to create normal user")

	userRequest := models.UserRegisterRequest{
		Username: "newuser_" + uuid.New().String()[:8],
		Password: "newpassword",
	}

	_, statusCode, err := userService.UserRegister(ctx, userRequest, normalUser.UUID)
	assert.Error(t, err, "Expected error when non-admin tries to register user")
	assert.Equal(t, 403, statusCode, "Expected status code 403")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, normalUser)
	})
}

func TestUserDeleteNonExistentUser(t *testing.T) {
	ctx := context.Background()

	nonExistentUUID := uuid.New()
	_, statusCode, err := userService.UserDelete(ctx, nonExistentUUID, uuid.Nil)
	assert.Error(t, err, "Expected error for non-existent user deletion")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestUserDeleteAsNonAdminOtherUser(t *testing.T) {
	ctx := context.Background()

	user1 := &models.User{
		UUID:     uuid.New(),
		Username: "user1_delete_" + uuid.New().String()[:8],
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	user2 := &models.User{
		UUID:     uuid.New(),
		Username: "user2_delete_" + uuid.New().String()[:8],
		Password: "password456",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, user1)
	assert.NoError(t, err, "Failed to create user1")

	err = userRepository.UserCreate(ctx, user2)
	assert.NoError(t, err, "Failed to create user2")

	_, statusCode, err := userService.UserDelete(ctx, user1.UUID, user2.UUID)
	assert.Error(t, err, "Expected error when non-admin tries to delete other user")
	assert.Equal(t, 403, statusCode, "Expected status code 403")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, user1)
		userRepository.UserDelete(ctx, user2)
	})
}

func TestUserListNonAdminCannotList(t *testing.T) {
	ctx := context.Background()

	user := &models.User{
		UUID:     uuid.New(),
		Username: "normaluser_list_" + uuid.New().String()[:8],
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, user)
	assert.NoError(t, err, "Failed to create user")

	_, statusCode, err := userService.UserList(ctx, user.UUID, 50)
	assert.Error(t, err, "Expected error when non-admin tries to list users")
	assert.Equal(t, 403, statusCode, "Expected status code 403")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, user)
	})
}

func TestUserListWithDifferentLengths(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		UUID:     uuid.New(),
		Username: "admin_list_lengths_" + uuid.New().String()[:8],
		Password: "adminpass123",
		Role:     models.UserRoleAdmin,
	}

	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	users10, statusCode, err := userService.UserList(ctx, admin.UUID, 10)
	assert.NoError(t, err, "Failed to list users with length 10")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.LessOrEqual(t, len(users10), 10, "Expected at most 10 users")

	users5, statusCode, err := userService.UserList(ctx, admin.UUID, 5)
	assert.NoError(t, err, "Failed to list users with length 5")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.LessOrEqual(t, len(users5), 5, "Expected at most 5 users")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, admin)
	})
}

func TestUserLoginNonExistentUsername(t *testing.T) {
	ctx := context.Background()

	loginRequest := models.UserLoginRequest{
		Username: "nonexistent_username_" + uuid.New().String()[:16],
		Password: "password123",
	}

	_, _, statusCode, err := userService.UserLogin(ctx, loginRequest)
	assert.Error(t, err, "Expected error for non-existent username")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestUserUpdateRoleByAdmin(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		UUID:     uuid.New(),
		Username: "admin_role_update_" + uuid.New().String()[:8],
		Password: "adminpass123",
		Role:     models.UserRoleAdmin,
	}

	user := &models.User{
		UUID:     uuid.New(),
		Username: "user_role_update_" + uuid.New().String()[:8],
		Password: "userpass123",
		Role:     models.UserRoleUser,
	}

	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin")

	err = userRepository.UserCreate(ctx, user)
	assert.NoError(t, err, "Failed to create user")

	newRole := models.UserRoleAdmin
	updateRequest := &models.UserUpdateRequest{
		Username: user.Username,
		Password: "userpass123",
		Role:     &newRole,
	}

	updatedUser, statusCode, err := userService.UserUpdate(ctx, admin.UUID, updateRequest)
	assert.NoError(t, err, "Admin role update failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, newRole, updatedUser.Role, "Expected role to be updated to admin")

	t.Cleanup(func() {
		userRepository.UserDelete(ctx, admin)
		userRepository.UserDelete(ctx, updatedUser)
	})
}
