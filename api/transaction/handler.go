package transaction

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/KKGo-Software-engineering/workshop-summer/api/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Err struct {
	Message string `json:"message"`
}
type handler struct {
	flag config.FeatureFlag
	db   *sql.DB
}

func New(cfg config.FeatureFlag, db *sql.DB) *handler {
	return &handler{cfg, db}
}

const (
	cStmt = `INSERT INTO transaction (date, amount, category, transaction_type, note, image_url, spender_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	uStmt = `UPDATE transaction SET date = $1, amount = $2, category = $3, transaction_type = $4, note = $5, image_url = $6, spender_id = $7 WHERE id = $8 RETURNING id;`
)

func (h handler) Create(c echo.Context) error {
	if !h.flag.EnableCreateTransaction {
		return c.JSON(http.StatusForbidden, "create new transaction feature is disabled")
	}

	logger := mlog.L(c)
	ctx := c.Request().Context()

	var tranReq TransactionRequest
	if err := c.Bind(&tranReq); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid transaction request"})
	}

	//validate spender id from spender table

	//create transaction
	var lastInsertId int64
	err := h.db.QueryRowContext(
		ctx,
		cStmt,
		tranReq.Date, tranReq.Amount, tranReq.Category, tranReq.TransactionType, tranReq.Note, tranReq.ImageUrl, tranReq.SpenderID,
	).Scan(&lastInsertId)

	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	logger.Info("create successfully", zap.Int64("id", lastInsertId))
	return c.JSON(http.StatusCreated, TransactionResponse{
		ID:              lastInsertId,
		Date:            &tranReq.Date,
		Amount:          tranReq.Amount,
		Category:        tranReq.Category,
		TransactionType: tranReq.TransactionType,
		Note:            tranReq.Note,
		ImageUrl:        tranReq.ImageUrl,
	})
}

func (h handler) GetAll(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	// page, err := strconv.Atoi(c.URL.Query().Get("page"))
	// if err != nil || page < 0 {
	// 	page = 0 // Default to the first page if page parameter is invalid
	// }

	// pageSize, err := strconv.Atoi(c.URL.Query().Get("pageSize"))
	// if err != nil || pageSize <= 0 {
	// 	pageSize = 10 // Default page size
	// }

	rows, err := h.db.QueryContext(ctx, `SELECT id, date, amount, category, transaction_type, note, image_url FROM transaction`)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var tRs []TransactionResponse
	for rows.Next() {
		var tR TransactionResponse
		err := rows.Scan(&tR.ID, &tR.Date, &tR.Amount, &tR.Category, &tR.TransactionType, &tR.Note, &tR.ImageUrl)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		tRs = append(tRs, tR)
	}

	return c.JSON(http.StatusOK, echo.Map{"transactions": tRs})
}

func GetByExpenseId() {

}

func (h *handler) GetTransactionById(c echo.Context) error {
	// Retrieve spenderID as a string and convert to integer
	spenderIDStr := c.Param("id")
	spenderID, err := strconv.Atoi(spenderIDStr)
	ctx := c.Request().Context()

	// fmt.Print("id is ", spenderIDStr)
	if err != nil {
		// Return an error if conversion fails
		c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid spender ID"})
		return err
	}

	var transactions []TransactionResponse

	// Use the integer spenderID in the SQL query
	rows, err := h.db.QueryContext(ctx, `
        SELECT id, date, amount, category, transaction_type, note, image_url
        FROM transaction
        WHERE spender_id = $1`, spenderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, echo.Map{"error": "Database error"})
		return err
	}
	defer rows.Close()
	fmt.Println("Transsation ", rows)

	for rows.Next() {
		var t TransactionResponse
		if err := rows.Scan(&t.ID, &t.Date, &t.Amount, &t.Category, &t.TransactionType, &t.Note, &t.ImageUrl); err != nil {
			c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning database results"})
			fmt.Println("print t ", t)
			return err

		}
		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error iterating database results"})
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"transactions": transactions})
}

func (h handler) Update(c echo.Context) error {
	if !h.flag.EnableUpdateTransaction {
		return c.JSON(http.StatusForbidden, "update transaction feature is disabled")
	}
	id, err := strconv.Atoi(c.Param("id"))
	if id == 0 {
		return c.JSON(http.StatusBadRequest, Err{Message: "ID is required"})
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid ID"})
	}

	logger := mlog.L(c)
	ctx := c.Request().Context()

	var tranReq TransactionRequest
	if err := c.Bind(&tranReq); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid transaction request"})
	}
	var lastInsertId int64
	err = h.db.QueryRowContext(ctx, uStmt,
		tranReq.Date, tranReq.Amount, tranReq.Category, tranReq.TransactionType, tranReq.Note, tranReq.ImageUrl, tranReq.SpenderID,
	).Scan(&lastInsertId)
	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	logger.Info("update successfully", zap.Int64("id", lastInsertId))
	return c.JSON(http.StatusOK, TransactionResponse{
		ID:              lastInsertId,
		Date:            &tranReq.Date,
		Amount:          tranReq.Amount,
		Category:        tranReq.Category,
		TransactionType: tranReq.TransactionType,
		Note:            tranReq.Note,
		ImageUrl:        tranReq.ImageUrl,
	})
}
