package export

import (
	"encoding/csv"
	"strconv"
	"strings"

	"task177-isomix/internal/model"
)

// CSV returns a spreadsheet-friendly representation without locale-specific
// formatting, so decimal values remain machine readable.
func CSV(report model.Report) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write([]string{"report_id", "title", "endmember_id", "name", "version", "low", "high"}); err != nil {
		return "", err
	}
	for _, row := range rows(report) {
		if err := writer.Write([]string{report.ID, title(report), row.EndmemberID, row.Name, strconv.Itoa(row.Version), strconv.FormatFloat(row.Low, 'g', -1, 64), strconv.FormatFloat(row.High, 'g', -1, 64)}); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}
