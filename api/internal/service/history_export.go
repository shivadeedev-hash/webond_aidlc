package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/webond/api/internal/domain"
)

// Export renders the filtered history as an .xlsx workbook (US-028 Add-on).
// Rejects with EXPORT_TOO_LARGE when the filtered set exceeds 10,000 rows
// (EX-14). Thai text is preserved (excelize writes UTF-8).
func (s *HistoryService) Export(ctx context.Context, userID uint64, f HistoryFilter) ([]byte, string, error) {
	// Pull all matching rows (bypass pagination) to size the export.
	f.Page = 1
	f.PageSize = maxExportRows + 1
	views, total, err := s.List(ctx, userID, f)
	if err != nil {
		return nil, "", err
	}
	if total > maxExportRows {
		return nil, "", domain.NewError(domain.CodeExportTooLarge,
			"Export exceeds 10,000 rows; narrow the range or use a monthly statement")
	}

	xl := excelize.NewFile()
	defer func() { _ = xl.Close() }()
	const sheet = "Transactions"
	idx, err := xl.NewSheet(sheet)
	if err != nil {
		return nil, "", fmt.Errorf("new sheet: %w", err)
	}
	xl.SetActiveSheet(idx)
	_ = xl.DeleteSheet("Sheet1")

	headers := []string{"Reference", "Date/Time", "Type", "Asset Type", "Amount (THB)", "Quantity", "Price", "Net (THB)", "Status"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xl.SetCellValue(sheet, cell, h)
	}
	for r, v := range views {
		row := r + 2
		vals := []any{v.TxnRef, v.OccurredAt, v.Type, v.AssetType, v.Amount, v.Quantity, v.Price, v.NetAmount, v.Status}
		for c, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			_ = xl.SetCellValue(sheet, cell, val)
		}
	}

	var buf bytes.Buffer
	if err := xl.Write(&buf); err != nil {
		return nil, "", fmt.Errorf("write xlsx: %w", err)
	}
	filename := fmt.Sprintf("webond-transactions-%s.xlsx", s.clock.Now().Format("20060102"))
	return buf.Bytes(), filename, nil
}
