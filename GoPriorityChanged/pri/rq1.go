package pri

import (
	"GoPriorityChanged/issue"
	"GoPriorityChanged/pro"
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"sort"
	"strings"
	"time"
)

type BugCount struct {
	Project                   string  `csv:"Project"`
	BugCount                  int     `csv:"BugCount"`
	BugOfPriorityChangedCount int     `csv:"BugOfPriorityChangedCount"`
	Percentage                float64 `csv:"Percentage"`
}

func RQ1(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	bugCounts := BugCounts(db, projects)
	sort.Slice(bugCounts[:len(bugCounts)-1], func(i, j int) bool {
		c := strings.Compare(bugCounts[i].Project, bugCounts[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	WriteBugCount(path, bugCounts)
	fmt.Println("RQ1耗时：", time.Since(start).String())
	Wg.Done()
}

// WriteBugCount 将ProjectBugCount结构体写入CSV文件
func WriteBugCount(path string, projectBugCount []BugCount) {
	err := csvtag.DumpToFile(projectBugCount, path)
	if err != nil {
		return
	}
}

// BugCounts 获取所有项目的BugCount情况
func BugCounts(db *gorm.DB, projects []string) []BugCount {
	bugCounts := make([]BugCount, 0)
	hadoop := BugCount{}
	hadoop.Project = "Hadoop"
	for _, project := range projects {
		bugCount := BugCountByProject(db, project)
		key := bugCount.Project
		if key == "hadoop" {
			hadoop.BugCount += bugCount.BugCount
			hadoop.BugOfPriorityChangedCount += bugCount.BugOfPriorityChangedCount
		} else {
			bugCounts = append(bugCounts, bugCount)
		}
	}
	f := float64(hadoop.BugOfPriorityChangedCount) / float64(hadoop.BugCount)
	hadoop.Percentage = RoundN(f*100, 1)
	bugCounts = append(bugCounts, hadoop)
	bc := BugCount{}
	bc.Project = "All projects"
	for _, bugCount := range bugCounts {
		bc.BugCount += bugCount.BugCount
		bc.BugOfPriorityChangedCount += bugCount.BugOfPriorityChangedCount
	}
	f = float64(bc.BugOfPriorityChangedCount) / float64(bc.BugCount)
	bc.Percentage = RoundN(f*100, 1)
	bugCounts = append(bugCounts, bc)
	return bugCounts
}

// BugCountByProject 获取该项目的BugCount
func BugCountByProject(db *gorm.DB, project string) BugCount {
	records := QueryBugRecordsByKeyPrefix(db, project)
	uniqueMap := make(map[string]bool)
	for _, record := range records {
		key := record.Key
		if _, ok := uniqueMap[key]; !ok {
			uniqueMap[key] = true
		}
	}
	count := issue.QueryProjectBugCount(db, project)
	bc := BugCount{
		Project:                   pro.GetNameByKey(project),
		BugCount:                  count,
		BugOfPriorityChangedCount: len(uniqueMap),
		Percentage:                RoundN(100*float64(len(uniqueMap))/float64(count), 1),
	}
	return bc
}

// RoundN 四舍五入保留n位小数
func RoundN(f float64, n int) float64 {
	f2, _ := decimal.NewFromFloat(f).Round(int32(n)).Float64()
	return f2
}
