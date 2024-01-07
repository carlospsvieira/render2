package handlers

import (
	"boss-payback/internal/api/auth"
	"boss-payback/internal/database"
	"boss-payback/internal/database/models"
	"boss-payback/pkg/helpers"
	"boss-payback/pkg/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Register(c *fiber.Ctx) error {
	var user models.User

	if err := utils.ParseRequestBody(c, &user); err != nil {
		return err
	}

	if user.Username == "" && user.Email == "" {
		return utils.HandleErrorResponse(c, fiber.StatusBadRequest, "Fields empty or missing")
	}

	if !helpers.ValidatePassword(user.Password) {
		return utils.HandleErrorResponse(c, fiber.StatusBadRequest, "Invalid password")
	}

	hashedPassword, err := helpers.HashPassword(user.Password)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password")
	}

	user.Password = string(hashedPassword)

	database.DB.Db.Create(&user)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"username": user.Username,
			"email":    user.Email,
			"roleId":   user.RoleID,
		},
		"message": fmt.Sprintf("%s was created!", user.Username),
	})
}

func Login(c *fiber.Ctx) error {
	var userRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := utils.ParseRequestBody(c, &userRequest); err != nil {
		return err
	}

	user, err := helpers.FindUser(userRequest.Username, userRequest.Password)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	if err := database.DB.Db.Model(&user).Where("id = ?", user.ID).Update("logged_in", true).Error; err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	token, err := auth.CreateToken(user.Username)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusInternalServerError, "Error generating token")
	}

	fmt.Println(token)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"username": user.Username,
			"roleId":   user.RoleID,
			"token":    token,
			"loggedIn": true,
		},
		"message": fmt.Sprintf("%s logged in successfully!", user.Username),
	})
}

func UpdateUsername(c *fiber.Ctx) error {
	var userRequest struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		NewUsername string `json:"newUsername"`
	}

	if err := utils.ParseRequestBody(c, &userRequest); err != nil {
		return err
	}

	if userRequest.NewUsername == "" {
		return utils.HandleErrorResponse(c, fiber.StatusBadRequest, "New username cannot be empty")
	}

	user, err := helpers.FindUser(userRequest.Username, userRequest.Password)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	if err := database.DB.Db.Model(&user).Where("id = ?", user.ID).Update("username", userRequest.NewUsername).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON("Username updated!")
}

func UpdatePassword(c *fiber.Ctx) error {
	var userRequest struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		NewPassword string `json:"newPassword"`
	}

	if err := utils.ParseRequestBody(c, &userRequest); err != nil {
		return err
	}

	user, err := helpers.FindUser(userRequest.Username, userRequest.Password)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	if !helpers.ValidatePassword(userRequest.Password) {
		return utils.HandleErrorResponse(c, fiber.StatusBadRequest, "New password does not meet criteria")
	}

	hashedPassword, err := helpers.HashPassword(userRequest.NewPassword)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password")
	}

	if err := database.DB.Db.Model(&user).Where("id = ?", user.ID).Update("password", hashedPassword).Error; err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).SendString("Password updated!")
}

func UpdateUserRole(c *fiber.Ctx) error {
	var userRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		RoleId   uint   `json:"roleId"`
	}

	if err := utils.ParseRequestBody(c, &userRequest); err != nil {
		return err
	}

	user, err := helpers.FindUser(userRequest.Username, userRequest.Password)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	if err := database.DB.Db.Model(&user).Where("id = ?", user.ID).Update("role_id", userRequest.RoleId).Error; err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).SendString("User's role updated!")
}

func GetUsersByRole(c *fiber.Ctx) error {
	var userRequest struct {
		RoleId uint `json:"roleId"`
	}

	type userResponse struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		RoleId   uint   `json:"roleId"`
	}

	if err := utils.ParseRequestBody(c, &userRequest); err != nil {
		return err
	}

	var users []models.User
	if err := database.DB.Db.Where("role_id = ?", userRequest.RoleId).Find(&users).Error; err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	userCh := make(chan userResponse)
	defer close(userCh)

	for _, user := range users {
		go func(u models.User) {
			userCh <- userResponse{
				Username: u.Username,
				Email:    u.Email,
				RoleId:   u.RoleID,
			}
		}(user)
	}

	var response []userResponse
	for range users {
		user := <-userCh
		response = append(response, user)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"users": response,
		},
		"message": fmt.Sprintf("Successfully fetched all users with role id %d", userRequest.RoleId),
	})
}

func DeleteUser(c *fiber.Ctx) error {
	var userRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := utils.ParseRequestBody(c, &userRequest); err != nil {
		return err
	}

	user, err := helpers.FindUser(userRequest.Username, userRequest.Password)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	if err := database.DB.Db.Unscoped().Delete(&user).Error; err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("%s was deleted!", user.Username),
	})
}
