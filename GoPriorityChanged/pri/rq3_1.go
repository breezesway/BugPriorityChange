package pri

import (
	"GoPriorityChanged/pro"
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"gorm.io/gorm"
	"sort"
	"strings"
	"time"
)

type ChangeCount struct {
	Project       string  `csv:"Project"`
	Total         int     `csv:"Total"`
	C1            int     `csv:"C1"`
	C1Percent     float64 `csv:"C1Percent"`
	C2            int     `csv:"C2"`
	C2Percent     float64 `csv:"C2Percent"`
	C3            int     `csv:"C3"`
	C3Percent     float64 `csv:"C3Percent"`
	C3Plus        int     `csv:"C3Plus"`
	C3PlusPercent float64 `csv:"C3PlusPercent"`
}

func RQ3_1(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	changeCounts := ChangeCounts(db, projects)
	sort.Slice(changeCounts[:len(changeCounts)-1], func(i, j int) bool {
		c := strings.Compare(changeCounts[i].Project, changeCounts[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	WriteChangeCounts(path, changeCounts)
	fmt.Println("RQ3_1耗时:", time.Since(start).String())
	Wg.Done()
}

// WriteChangeCounts 将所有项目的ChangeCount写入CSV文件
func WriteChangeCounts(path string, changeCounts []ChangeCount) {
	err := csvtag.DumpToFile(changeCounts, path)
	if err != nil {
		fmt.Println("写入出错!")
		return
	}
}

// ChangeCounts 获取所有项目的ChangeCount分布情况
func ChangeCounts(db *gorm.DB, projects []string) []ChangeCount {
	changeCounts := make([]ChangeCount, 0)
	hadoop := ChangeCount{}
	hadoop.Project = "Hadoop"
	for _, project := range projects {
		changeCount := ChangeCountByProject(db, project)
		if project == "HADOOP" || project == "HDFS" || project == "MAPREDUCE" || project == "YARN" {
			ccAccumulate(&hadoop, &changeCount)
		} else {
			changeCounts = append(changeCounts, changeCount)
		}
	}
	ccPercentage(&hadoop)
	changeCounts = append(changeCounts, hadoop)
	cc := ChangeCount{}
	cc.Project = "All projects"
	for _, changeCount := range changeCounts {
		ccAccumulate(&cc, &changeCount)
	}
	ccPercentage(&cc)
	changeCounts = append(changeCounts, cc)
	return changeCounts
}

// ChangeCountByProject 计算该项目的优先级变更次数分布
func ChangeCountByProject(db *gorm.DB, project string) ChangeCount {
	records := QueryBugRecordsByKeyPrefix(db, project)
	m := make(map[string]int)
	for _, record := range records {
		key := record.Key
		if count, ok := m[key]; !ok {
			m[key] = 1
		} else {
			m[key] = count + 1
		}
	}
	cc := ChangeCount{}
	cc.Project = pro.GetNameByKey(project)
	cc.Total = len(m)
	for _, v := range m {
		switch v {
		case 1:
			cc.C1++
		case 2:
			cc.C2++
		case 3:
			cc.C3++
		default:
			cc.C3Plus++
		}
	}
	ccPercentage(&cc)
	return cc
}

func ccAccumulate(cc, changeCount *ChangeCount) {
	cc.Total += changeCount.Total
	cc.C1 += changeCount.C1
	cc.C2 += changeCount.C2
	cc.C3 += changeCount.C3
	cc.C3Plus += changeCount.C3Plus
}

func ccPercentage(cc *ChangeCount) {
	f := float64(cc.Total)
	cc.C1Percent = RoundN(100*float64(cc.C1)/f, 1)
	cc.C2Percent = RoundN(100*float64(cc.C2)/f, 1)
	cc.C3Percent = RoundN(100*float64(cc.C3)/f, 1)
	cc.C3PlusPercent = RoundN(100*float64(cc.C3Plus)/f, 1)
}
