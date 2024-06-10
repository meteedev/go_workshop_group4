package transaction

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetSummaryTransaction(t *testing.T) {
	t.Run("get summary transaction successfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/spenders/123/transactions/summary", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("failed to create sqlmock: %s", err)
		}
		
		defer db.Close()

		
		
		spender_id :=123
		mock.ExpectQuery(summary_stmt).
        WithArgs(spender_id).
        WillReturnRows(sqlmock.NewRows([]string{"total_amount", "tran_type"}).
            AddRow(500.00, "income").
            AddRow(300.00, "expense"))

		cfg := config.FeatureFlag{EnableCreateSpender: true}

		h := New(cfg, db)

		summaryTrans , err := h.getSummaryTransaction(c.Request().Context(),spender_id)

		assert.NoError(t, err)
		expectedSummary := []SummaryTransaction{
			{TransactionType: "income", TotalAmount: 500.00},
			{TransactionType: "expense", TotalAmount: 300.00},
		}
		assert.Equal(t, expectedSummary, summaryTrans)
	})


}


func TestGetSpenderSummary(t *testing.T) {
	// Step 1: Create a new Echo instance for testing
	e := echo.New()

	// Step 2: Create a new SQL mock


	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

	if err != nil {
		t.Fatalf("failed to create sqlmock: %s", err)
	}
	defer db.Close()

	// Step 3: Defer closing the mock database connection
	defer mock.ExpectationsWereMet()

	spender_id :=123
	mock.ExpectQuery(summary_stmt).
	WithArgs(spender_id).
	WillReturnRows(sqlmock.NewRows([]string{"total_amount", "tran_type"}).
		AddRow(500.00, "income").
		AddRow(300.00, "expense"))

	// Step 4: Create a handler instance with the mocked database
	h := handler{
		db: db,
	}



	// Step 6: Test case: Valid spender ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/spenders/123/transactions/summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123")
	err = h.GetSpenderSummary(c)

	// Step 7: Assert the response status code
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Step 8: Assert the response body
	expectedResponseBody := `{
		"summary": {
			"total_income": 500,
			"total_expenses": 300,
			"current_balance": 200
		}
	}`
	assert.JSONEq(t, expectedResponseBody, rec.Body.String())
}


func TestCalculateSummary(t *testing.T) {
	t.Run("Only income transactions", func(t *testing.T) {
		// Test case 1: Only income transactions
		transactions := []SummaryTransaction{
			{TransactionType: "income", TotalAmount: 500.0},
			{TransactionType: "income", TotalAmount: 300.0},
		}
		expectedSummary := SummaryResponse{
			Summary: struct {
				TotalIncome    float64 `json:"total_income"`
				TotalExpenses  float64 `json:"total_expenses"`
				CurrentBalance float64 `json:"current_balance"`
			}{
				TotalIncome:    800.0,
				TotalExpenses:  0.0,
				CurrentBalance: 800.0,
			},
		}
		assert.Equal(t, expectedSummary, calculateSummary(transactions))
	})

	t.Run("Only expense transactions", func(t *testing.T) {
		// Test case 2: Only expense transactions
		transactions := []SummaryTransaction{
			{TransactionType: "expense", TotalAmount: 200.0},
			{TransactionType: "expense", TotalAmount: 100.0},
		}
		expectedSummary := SummaryResponse{
			Summary: struct {
				TotalIncome    float64 `json:"total_income"`
				TotalExpenses  float64 `json:"total_expenses"`
				CurrentBalance float64 `json:"current_balance"`
			}{
				TotalIncome:    0.0,
				TotalExpenses:  300.0,
				CurrentBalance: -300.0,
			},
		}
		assert.Equal(t, expectedSummary, calculateSummary(transactions))
	})

	t.Run("Mixed income and expense transactions", func(t *testing.T) {
		// Test case 3: Mixed income and expense transactions
		transactions := []SummaryTransaction{
			{TransactionType: "income", TotalAmount: 1000.0},
			{TransactionType: "expense", TotalAmount: 300.0},
			{TransactionType: "income", TotalAmount: 500.0},
		}
		expectedSummary := SummaryResponse{
			Summary: struct {
				TotalIncome    float64 `json:"total_income"`
				TotalExpenses  float64 `json:"total_expenses"`
				CurrentBalance float64 `json:"current_balance"`
			}{
				TotalIncome:    1500.0,
				TotalExpenses:  300.0,
				CurrentBalance: 1200.0,
			},
		}
		assert.Equal(t, expectedSummary, calculateSummary(transactions))
	})
}
