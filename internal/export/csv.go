package export

import (
	"encoding/csv"
	"strconv"
	"strings"

	"task177-isomix/internal/model"
)

// CSV 返回电子表格友好的导出（不含本地化格式，使小数值保持机器可读）。
// 表头列顺序与 ColumnNames / rows 保持一致，端元行按 endmember_id 升序输出。
func CSV(report model.Report) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write(columns); err != nil {
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
