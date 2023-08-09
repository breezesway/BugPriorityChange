package main

import (
	"GoPriorityChanged/issue"
	_ "GoPriorityChanged/issue"
	"GoPriorityChanged/pri"
	"GoPriorityChanged/pro"
	"fmt"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

type CSBug struct {
	Key             string    `gorm:"column:Key"`
	Project         string    `gorm:"column:Project"`
	IssueType       string    `gorm:"column:IssueType"`
	Status          string    `gorm:"column:Status"`
	Priority        string    `gorm:"column:Priority"`
	Resolution      string    `gorm:"column:Resolution"`
	Created         time.Time `gorm:"column:Created"`
	ChangePhase     string    `gorm:"column:ChangePhase"`
	ChangeDirection string    `gorm:"column:ChangeDirection"`
	PCTime          string    `gorm:"column:PCTime"`
}

func handleCSBug(db *gorm.DB) {
	fileMap := pri.GetProComFileMap("E:\\PriorityChange\\ProjectCommit")
	total := 0
	for k := range fileMap {
		keys, _ := pro.GetKeyByName(k)
		for _, key := range keys {
			csbugs := make([]CSBug, 0)
			issues := issue.QueryBugsByProject(db, key)
			fmt.Println(key, len(issues))
			total += len(issues)
			for _, i := range issues {
				c := CSBug{}
				c.Key = i.Key
				c.Project = k
				c.IssueType = "Bug"
				c.Status = i.Status
				c.Priority = i.Priority
				c.Resolution = i.Resolution
				c.Created = pri.StringToDateTime(i.Created)
				if pcs, ok := pri.RecordsMap[i.Key]; ok {
					c.ChangePhase = "A"
					c.ChangeDirection = "Increase"
					c.PCTime = pcs[0].Created
				} else {
					c.ChangePhase = ""
					c.ChangeDirection = ""
					//c.PCTime = ""
				}
				csbugs = append(csbugs, c)
			}
			i := 0
			for i < len(csbugs) {
				var tempCSbugs []CSBug
				if i+500 < len(csbugs) {
					tempCSbugs = csbugs[i : i+500]
				} else {
					tempCSbugs = csbugs[i:]
				}
				i += 500
				db.Table("csbug").Clauses(clause.OnConflict{DoNothing: true}).Create(&tempCSbugs)
			}
		}
	}
}

func calCSBugAve() {
	fileMap := pri.GetProComFileMap("E:\\PriorityChange\\ave")
	sl := make([]float64, 10)
	for _, files := range fileMap {
		for _, f := range files {
			xlsx, _ := excelize.OpenFile(f)

			value, _ := xlsx.GetCellValue("SEKEResults", "D14")
			notCount, _ := strconv.ParseFloat(value, 64)
			value, _ = xlsx.GetCellValue("SEKEResults", "E14")
			aveCount, _ := strconv.ParseFloat(value, 64)
			sl[8] += notCount
			sl[9] += aveCount

			value, _ = xlsx.GetCellValue("SEKEResults", "B15")
			v, _ := strconv.ParseFloat(value, 64)
			sl[0] += v * notCount
			value, _ = xlsx.GetCellValue("SEKEResults", "C15")
			v, _ = strconv.ParseFloat(value, 64)
			sl[1] += v * aveCount

			value, _ = xlsx.GetCellValue("SEKEResults", "B17")
			v, _ = strconv.ParseFloat(value, 64)
			sl[2] += v * notCount
			value, _ = xlsx.GetCellValue("SEKEResults", "C17")
			v, _ = strconv.ParseFloat(value, 64)
			sl[3] += v * aveCount

			value, _ = xlsx.GetCellValue("SEKEResults", "B19")
			v, _ = strconv.ParseFloat(value, 64)
			sl[4] += v * notCount
			value, _ = xlsx.GetCellValue("SEKEResults", "C19")
			v, _ = strconv.ParseFloat(value, 64)
			sl[5] += v * aveCount

			value, _ = xlsx.GetCellValue("SEKEResults", "B21")
			v, _ = strconv.ParseFloat(value, 64)
			sl[6] += v * notCount
			value, _ = xlsx.GetCellValue("SEKEResults", "C21")
			v, _ = strconv.ParseFloat(value, 64)
			sl[7] += v * aveCount
		}
	}
	for i := 0; i < 8; i += 2 {
		sl[i] = sl[i] / sl[8]
		sl[i+1] = sl[i+1] / sl[9]
	}
	for i := 0; i < 10; i += 2 {
		fmt.Println(sl[i], sl[i+1])
	}
}
