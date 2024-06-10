package transaction

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
) 


type SummaryTransaction struct{
	TransactionType string
	TotalAmount	float64
}


type SummaryResponse struct {
	Summary struct {
		TotalIncome    float64 `json:"total_income"`
		TotalExpenses  float64 `json:"total_expenses"`
		CurrentBalance float64 `json:"current_balance"`
	} `json:"summary"`
}



const (
	summary_stmt  = `SELECT sum(amount) as total_amount, transaction_type as tran_type FROM transaction WHERE spender_id = $1 group by spender_id ,transaction_type`
)

func (h handler) GetSpenderSummary(c echo.Context) error{
	id , err := strconv.Atoi(c.Param("id"))
	if err != nil {
        return c.JSON(http.StatusBadRequest, "invalid spender id")
    }

	ctx := c.Request().Context()
	summaryTran, err := h.getSummaryTransaction(ctx,id)

	if err != nil {
        fmt.Println(err.Error())
        return c.JSON(http.StatusInternalServerError, "getSpenderSummary error")
    }

    if len(summaryTran) == 0{
        return c.JSON(http.StatusNotFound, " transaction not found")
    }

	return c.JSON(http.StatusOK, calculateSummary(summaryTran))
}

func (h handler) getSummaryTransaction(ctx context.Context, id int)(summaryTrans []SummaryTransaction , err error){
	rows, err := h.db.QueryContext(ctx, summary_stmt,id)
	if err != nil {	
		return nil,err
	}
	defer rows.Close()

	var s []SummaryTransaction

    for rows.Next() {
        var totalAmount float64
        var transactionType string

        if err := rows.Scan(&totalAmount, &transactionType); err != nil {
            return nil, err
        }

        s = append(s, SummaryTransaction{
            TransactionType: transactionType,
            TotalAmount:     totalAmount,
        })
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return s, nil

}

func calculateSummary(transactions []SummaryTransaction) SummaryResponse {
    // Initialize variables for total income and total expenses
    totalIncome := 0.0
    totalExpenses := 0.0

    // Loop through transactions
    for _, transaction := range transactions {
        if transaction.TransactionType == "income" {
            totalIncome += transaction.TotalAmount
        } else if transaction.TransactionType == "expense" {
            totalExpenses += transaction.TotalAmount
        }
    }

    // Calculate current balance
    currentBalance := totalIncome - totalExpenses

    // Create and return summary response
    return SummaryResponse{
        Summary: struct {
            TotalIncome    float64 `json:"total_income"`
            TotalExpenses  float64 `json:"total_expenses"`
            CurrentBalance float64 `json:"current_balance"`
        }{
            TotalIncome:    totalIncome,
            TotalExpenses:  totalExpenses,
            CurrentBalance: currentBalance,
        },
    }
}


