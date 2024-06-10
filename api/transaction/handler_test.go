package transaction

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateTransaction(t *testing.T) {

	t.Run("create transaction fail when bad request body", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{ bad request body }`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		cfg := config.FeatureFlag{EnableCreateTransaction: true}

		h := New(cfg, nil)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid transaction request")
	})

	t.Run("create transaction success when feature toggle is enable", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", io.NopCloser(strings.NewReader(`{
			"date": "2024-04-30T09:00:00.000Z",
			"amount": 1000,
			"category": "Food",
			"transaction_type": "expense",
			"note": "Lunch",
			"image_url": "https://example.com/image1.jpg",
			"spender_id": 1
		}`)))

		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("api/v1/transactions")

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		layout := "2006-01-02T15:04:05.000Z"
		str := "2024-04-30T09:00:00.000Z"
		date, err := time.Parse(layout, str)
		if err != nil {
			fmt.Println(err)
		}

		tr := TransactionRequest{
			Date:            date,
			Amount:          1000,
			Category:        "Food",
			TransactionType: "expense",
			Note:            "Lunch",
			ImageUrl:        "https://example.com/image1.jpg",
			SpenderID:       1,
		}

		row := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery(cStmt).WithArgs(tr.Date, tr.Amount, tr.Category, tr.TransactionType, tr.Note, tr.ImageUrl, tr.SpenderID).WillReturnRows(row)
		cfg := config.FeatureFlag{EnableCreateTransaction: true}

		h := New(cfg, db)

		err = h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var got TransactionResponse
		err = rec.Result().Body.Close()
		if err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
		assert.Equal(t, TransactionResponse{
			ID:              1,
			Date:            &tr.Date,
			Amount:          tr.Amount,
			Category:        tr.Category,
			TransactionType: tr.TransactionType,
			Note:            tr.Note,
			ImageUrl:        tr.ImageUrl,
		}, got)
	})
	t.Run("create transaction fail when feature toggle is disable", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{
			"date": "2024-04-30T09:00:00.000Z",
			"amount": 1000,
			"category": "Food",
			"transaction_type": "expense",
			"note": "Lunch",
			"image_url": "https://example.com/image1.jpg",
			"spender_id": 1
		}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		cfg := config.FeatureFlag{EnableCreateTransaction: false}

		h := New(cfg, nil)
		err := h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
	t.Run("create transaction fail when query row error", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{
			"date": "2024-04-30T09:00:00.000Z",
			"amount": 1000,
			"category": "Food",
			"transaction_type": "expense",
			"note": "Lunch",
			"image_url": "https://example.com/image1.jpg",
			"spender_id": 1
		}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		cfg := config.FeatureFlag{EnableCreateTransaction: true}

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		layout := "2006-01-02T15:04:05.000Z"
		str := "2024-04-30T09:00:00.000Z"
		date, err := time.Parse(layout, str)
		if err != nil {
			fmt.Println(err)
		}

		tr := TransactionRequest{
			Date:            date,
			Amount:          1000,
			Category:        "Food",
			TransactionType: "expense",
			Note:            "Lunch",
			ImageUrl:        "https://example.com/image1.jpg",
			SpenderID:       1,
		}

		mock.ExpectQuery(cStmt).WithArgs(tr.Date, tr.Amount, tr.Category, tr.TransactionType, tr.Note, tr.ImageUrl, tr.SpenderID).WillReturnError(fmt.Errorf("query row error"))
		h := New(cfg, db)
		err = h.Create(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestUpdateTransaction(t *testing.T) {
	t.Run("update transaction fail when bad request body", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{ bad request body }`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("api/v1/transactions/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		cfg := config.FeatureFlag{EnableUpdateTransaction: true}

		h := New(cfg, nil)
		err := h.Update(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid transaction request")
	})

	t.Run("update transaction success when feature toggle is enable", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPut, "/", io.NopCloser(strings.NewReader(`{
			"date": "2024-04-30T09:00:00.000Z",
			"amount": 1000,
			"category": "Food",
			"transaction_type": "expense",
			"note": "Lunch",
			"image_url": "https://example.com/image1.jpg",
			"spender_id": 1
		}`)))

		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("api/v1/transactions/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		layout := "2006-01-02T15:04:05.000Z"
		str := "2024-04-30T09:00:00.000Z"
		date, err := time.Parse(layout, str)
		if err != nil {
			fmt.Println(err)
		}

		tr := TransactionRequest{
			Date:            date,
			Amount:          1000,
			Category:        "Food",
			TransactionType: "expense",
			Note:            "Lunch",
			ImageUrl:        "https://example.com/image1.jpg",
			SpenderID:       1,
		}

		row := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery(uStmt).WithArgs(tr.Date, tr.Amount, tr.Category, tr.TransactionType, tr.Note, tr.ImageUrl, tr.SpenderID).WillReturnRows(row)
		cfg := config.FeatureFlag{EnableUpdateTransaction: true}

		h := New(cfg, db)

		err = h.Update(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var got TransactionResponse
		err = rec.Result().Body.Close()
		if err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
		assert.Equal(t, TransactionResponse{
			ID:              1,
			Date:            &tr.Date,
			Amount:          tr.Amount,
			Category:        tr.Category,
			TransactionType: tr.TransactionType,
			Note:            tr.Note,
			ImageUrl:        tr.ImageUrl,
		}, got)
	})
}

func TestGetTransactionById(t *testing.T) {
	e := echo.New()
	defer e.Close()

	t.Run("successfully retrieve transactions by spender ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/transactions/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		defer db.Close()

		// Convert string to time.Time
		date1, _ := time.Parse(time.RFC3339, "2022-01-01T12:00:00Z")
		date2, _ := time.Parse(time.RFC3339, "2022-01-02T12:00:00Z")

		rows := sqlmock.NewRows([]string{"id", "date", "amount", "category", "transaction_type", "note", "image_url"}).
			AddRow(1, date1, 100.00, "groceries", "expense", "Weekly groceries", "http://example.com/receipt1.jpg").
			AddRow(2, date2, 150.00, "electronics", "expense", "Gadget purchase", "http://example.com/receipt2.jpg")
		mock.ExpectQuery(`SELECT id, date, amount, category, transaction_type, note, image_url FROM transaction WHERE spender_id = \$1`).WithArgs(1).WillReturnRows(rows)

		h := New(config.FeatureFlag{}, db)
		if assert.NoError(t, h.GetTransactionById(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, `{"transactions":[{"id":1,"date":"2022-01-01T12:00:00Z","amount":100.00,"category":"groceries","transaction_type":"expense","note":"Weekly groceries","image_url":"http://example.com/receipt1.jpg"},{"id":2,"date":"2022-01-02T12:00:00Z","amount":150.00,"category":"electronics","transaction_type":"expense","note":"Gadget purchase","image_url":"http://example.com/receipt2.jpg"}]}`, rec.Body.String())
		}
	})

	t.Run("database error during transaction retrieval", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/transactions/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		defer db.Close()

		// Configure the mock to return an error for the query
		mock.ExpectQuery(`SELECT id, date, amount, category, transaction_type, note, image_url FROM transaction WHERE spender_id = \$1`).WithArgs(1).WillReturnError(assert.AnError)

		h := New(config.FeatureFlag{}, db)
		err := h.GetTransactionById(c)

		// Test the error handling and response
		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Database error")
	})
}
func TestGetAllTransaction(t *testing.T) {
	t.Run("get all transaction succesfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		// Convert string to time.Time
		date1, _ := time.Parse(time.RFC3339, "2024-04-30T09:00:00Z")
		date2, _ := time.Parse(time.RFC3339, "2024-04-29T19:00:00Z")

		rows := sqlmock.NewRows([]string{"id", "date", "amount", "category", "transaction_type", "note", "image_url"}).
			AddRow(1, date1, 1000.00, "Food", "expense", "Lunch", "https://example.com/image1.jpg").
			AddRow(2, date2, 2000.00, "Transport", "income", "Salary", "https://example.com/image2.jpg")
		mock.ExpectQuery(`SELECT id, date, amount, category, transaction_type, note, image_url FROM transaction`).WillReturnRows(rows)

		h := New(config.FeatureFlag{}, db)
		err := h.GetAll(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"transactions":[{"id":1,"date":"2024-04-30T09:00:00Z","amount":1000.00,"category":"Food","transaction_type":"expense","note":"Lunch","image_url":"https://example.com/image1.jpg"},{"id":2,"date":"2024-04-29T19:00:00Z","amount":2000.00,"category":"Transport","transaction_type":"income","note":"Salary","image_url":"https://example.com/image2.jpg"}]}`, rec.Body.String())
	})

	t.Run("get all transaction failed on database", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectQuery(`SELECT id, date, amount, category, transaction_type, note, image_url FROM transaction`).WillReturnError(assert.AnError)

		h := New(config.FeatureFlag{}, db)
		err := h.GetAll(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
