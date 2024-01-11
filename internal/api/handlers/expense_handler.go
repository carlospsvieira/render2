package handlers

import (
	"boss-payback/internal/api/services"
	"boss-payback/internal/database/db_services"
	"boss-payback/internal/database/models"
	"boss-payback/pkg/helpers"
	"boss-payback/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateExpense(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	amount, userId, filePath, description, err := helpers.ExpenseForm(c, form)
	if err != nil {
		return utils.HandleErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	expense := models.Expense{
		UserID:       userId,
		Description:  description,
		Amount:       amount,
		ReceiptImage: filePath,
	}

	db_services.CreateExpenseInDB(c, &expense)

	return services.CreateExpenseResponse(c, &expense)
}

func UpdateExpenseAmount(c *fiber.Ctx) error {
	var request UpdateExpenseAmountRequest
	utils.ParseRequestBody(c, &request)

	if request.Amount == 0 {
		return utils.HandleErrorResponse(c, fiber.StatusBadRequest, "Amount cannot be missing or zero")
	}

	db_services.UpdateExpenseAmountInDB(c, request.ID, request.Amount)

	return services.UpdateExpenseAmountResponse(c, request.Amount, request.ID)
}

func UpdateExpenseDescription(c *fiber.Ctx) error {
	var request UpdateExpenseDescriptionRequest
	utils.ParseRequestBody(c, &request)

	db_services.UpdateExpenseDescriptionInDB(c, request.ID, request.Description)

	return services.UpdateExpenseDescriptionResponse(c, request.Description, request.ID)
}

func GetExpenses(c *fiber.Ctx) error {
	var expenses []models.Expense

	db_services.GetExpensesInDB(c, &expenses)

	return services.GetExpensesResponse(c, &expenses)
}

func GetExpensesByUser(c *fiber.Ctx) error {
	var request GetExpensesByUserRequest
	var expenses []models.Expense

	utils.ParseRequestBody(c, &request)

	db_services.GetExpensesByUserInDB(c, &expenses, request.UserID)

	return services.GetExpensesByUserResponse(c, &expenses, request.UserID)
}

func DeleteExpense(c *fiber.Ctx) error {
	var expense models.Expense
	utils.ParseRequestBody(c, &expense)

	db_services.DeleteExpenseInDB(c, &expense)

	return services.DeleteExpenseResponse(c, &expense)
}
