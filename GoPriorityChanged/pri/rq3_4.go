package pri

import (
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"gorm.io/gorm"
	"math"
	"strings"
	"time"
)

type ChangeType struct {
	Type   string `csv:"Type"`
	C1     int    `csv:"C1"`
	C2     int    `csv:"C2"`
	C3     int    `csv:"C3"`
	C3Plus int    `csv:"C3Plus"`
	All    int    `csv:"All"`
}

func RQ3_4_1(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	cts := ChangeTypes(db, projects)
	WriteChangeTypes(cts, path)
	fmt.Println("RQ3_4_1耗时:", time.Since(start).String())
	Wg.Done()
}

// WriteChangeTypes 将ChangeType写入CSV文件
func WriteChangeTypes(cts []ChangeType, path string) {
	err := csvtag.DumpToFile(cts, path)
	if err != nil {
		fmt.Println("写入出错")
		return
	}
}

// ChangeTypes 计算所有项目的变更类型次数（按变更次数来分）
func ChangeTypes(db *gorm.DB, projects []string) []ChangeType {
	cts := make([]ChangeType, 3)
	cts[0].Type = "Increase"
	cts[1].Type = "Decrease"
	cts[2].Type = "Restoration"
	for _, project := range projects {
		m := ChangeSequenceByProject(db, project)
		for _, s := range m {
			c := len(s) - 1
			if strings.Compare(s[:1], s[len(s)-1:]) > 0 {
				addByC(c, &cts[0])
			} else if strings.Compare(s[:1], s[len(s)-1:]) < 0 {
				addByC(c, &cts[1])
			} else {
				addByC(c, &cts[2])
			}
		}
	}
	return cts
}

//根据变更次数添加一次
func addByC(c int, changeType *ChangeType) {
	changeType.All++
	if c == 1 {
		changeType.C1++
	} else if c == 2 {
		changeType.C2++
	} else if c == 3 {
		changeType.C3++
	} else {
		changeType.C3Plus++
	}
}

type ChangeRange struct {
	Key  string  `csv:"Key"`
	Mean float64 `csv:"Mean"`
	Std  float64 `csv:"Std"`
}

func RQ3_4_2(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	crs := ChangeRanges(db, projects)
	WriteChangeRanges(crs, path)
	fmt.Println("RQ3_4_2耗时:", time.Since(start).String())
	Wg.Done()
}

// WriteChangeRanges 将ChangeRange写入CSV文件
func WriteChangeRanges(crs []ChangeRange, path string) {
	err := csvtag.DumpToFile(crs, path)
	if err != nil {
		fmt.Println("写入出错")
		return
	}
}

// ChangeRanges 计算所有项目的每个Bug的变更幅度
func ChangeRanges(db *gorm.DB, projects []string) []ChangeRange {
	crs := make([]ChangeRange, 0)
	for _, project := range projects {
		m := ChangeSequenceByProject(db, project)
		for k, s := range m {
			c := ChangeRange{}
			c.Key = k
			cr := CSToCR(s)
			c.Mean = mean(cr)
			c.Std = std(cr)
			crs = append(crs, c)
		}
	}
	return crs
}

// CSToCR 将变更序列转为换变更幅度
func CSToCR(cs string) []float64 {
	cr := make([]float64, len(cs)-1)
	for i := 0; i < len(cs)-1; i++ {
		v := cs[i+1] - cs[i]
		if v < 5 {
			cr[i] = float64(v)
		} else {
			cr[i] = 256 - float64(v)
		}
	}
	return cr
}

func mean(v []float64) float64 {
	var res float64 = 0
	var n = len(v)
	for i := 0; i < n; i++ {
		res += v[i]
	}
	return res / float64(n)
}

func variance(v []float64) float64 {
	n := len(v)
	if n == 1 {
		return 0
	}
	var res float64 = 0
	var m = mean(v)
	for i := 0; i < n; i++ {
		res += (v[i] - m) * (v[i] - m)
	}
	return res / float64(n-1)
}

func std(v []float64) float64 {
	return math.Sqrt(variance(v))
}
