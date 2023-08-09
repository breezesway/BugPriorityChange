package his

import (
	"gorm.io/gorm"
	"strings"
)

type History struct {
	His   RowHis
	Items []Item
}

type RowHis struct {
	Id       string `gorm:"column:id"`
	Author   string `gorm:"column:authordisplayname"`
	Created  string `gorm:"column:created"`
	IssueKey string `gorm:"column:issueKey"`
	RowItems string `gorm:"column:items"`
}

type Item struct {
	Field      string
	FieldType  string
	From       string
	FromString string
	To         string
	ToString   string
}

// QueryFieldModifiersByRowIds 根据原始Id数组查询并返回所有字段修改者集合
func QueryFieldModifiersByRowIds(db *gorm.DB, rowIds string) map[string]bool {
	histories := QueryHistoriesByRowIds(db, rowIds)
	fieldModifierMap := make(map[string]bool)
	for i := 0; i < len(histories); i++ {
		otherFiled := false
		for _, item := range histories[i].Items {
			if item.Field != "priority" {
				otherFiled = true
				break
			}
		}
		if otherFiled {
			fieldModifierMap[histories[i].His.Author] = true
		}
	}
	return fieldModifierMap
}

// QueryHistoriesByRowIds 根据原生的Id数组查询，并返回一组History记录
func QueryHistoriesByRowIds(db *gorm.DB, historyIds string) []History {
	historyIds = historyIds[1 : len(historyIds)-1]
	split := strings.Split(historyIds, ", ")
	histories := make([]History, 0)
	for _, s := range split {
		history := QueryHistoryById(db, s)
		histories = append(histories, history)
	}
	return histories
}

func QueryHistoryById(db *gorm.DB, id string) History {
	var history History
	var rowhis RowHis
	db.Table("history").
		Select("id", "authordisplayname", "created", "issueKey", "items").
		Where("id = ?", id).
		Find(&rowhis)
	history.His = rowhis
	parseRowItems(&history)
	return history
}

func parseRowItems(h *History) {
	h.Items = make([]Item, 0)
	row := h.His.RowItems
	rowItem := make([]string, 0)
	for {
		left := strings.Index(row, "(field=")
		right := strings.Index(row, "), History.Item(")
		if right == -1 {
			right = len(row) - 2
		}
		if left != -1 && right != -1 {
			rowItem = append(rowItem, row[left+1:right])
			row = row[right+1:]
		} else {
			break
		}
	}
	for _, r := range rowItem {
		item := Item{}
		left := -1
		right := -1
		var s string

		left = strings.Index(r, "field")
		right = strings.Index(r, ", fieldtype")
		item.Field = r[left+6 : right]
		r = r[right:]

		left = strings.Index(r, "fieldtype")
		right = strings.Index(r, ", from")
		item.FieldType = r[left+10 : right]
		r = r[right+1:]

		left = strings.Index(r, "from=")
		right = strings.Index(r, ", fromString=")
		s = r[left+5 : right]
		if s != "null" {
			item.From = s
		}
		r = r[right+1:]

		left = strings.Index(r, "fromString=")
		right = strings.Index(r, ", to=")
		s = r[left+11 : right]
		if s != "null" {
			item.FromString = s
		}
		r = r[right+1:]

		left = strings.Index(r, "to=")
		right = strings.Index(r, ", toString=")
		s = r[left+3 : right]
		if s != "null" {
			item.To = s
		}
		r = r[right+1:]

		left = strings.Index(r, "toString=")
		s = r[left+9:]
		if s != "null" {
			item.ToString = s
		}
		h.Items = append(h.Items, item)
	}
}
