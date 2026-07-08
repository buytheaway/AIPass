package service

import (
	"bytes"
	"context"

	"github.com/aipass/aipass/internal/repository"
	"github.com/xuri/excelize/v2"
)

type ReportService struct {
	repo *repository.Store
}

func (s *ReportService) AccessEventsXLSX(ctx context.Context) ([]byte, error) {
	events, err := s.repo.ListAccessEvents(ctx)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	sheet := "access_events"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"id", "user_id", "event_type", "decision", "reason", "scanner_id", "occurred_at"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}
	for row, event := range events {
		values := []any{event.ID.String(), event.UserID.String(), string(event.EventType), string(event.Decision), "", "", event.OccurredAt.Format("2006-01-02 15:04:05")}
		if event.Reason != nil {
			values[4] = *event.Reason
		}
		if event.ScannerID != nil {
			values[5] = *event.ScannerID
		}
		for col, value := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			_ = f.SetCellValue(sheet, cell, value)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ReportService) PaymentsXLSX(ctx context.Context) ([]byte, error) {
	payments, err := s.repo.ListPayments(ctx)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	sheet := "payments"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"id", "user_id", "subscription_id", "amount", "currency", "method", "status", "created_at"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}
	for row, payment := range payments {
		values := []any{
			payment.ID.String(), payment.UserID.String(), payment.SubscriptionID.String(),
			payment.Amount.StringFixedBank(2), payment.Currency, string(payment.Method), string(payment.Status),
			payment.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		for col, value := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			_ = f.SetCellValue(sheet, cell, value)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
