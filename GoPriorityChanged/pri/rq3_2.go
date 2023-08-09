package pri

import (
	"GoPriorityChanged/pro"
	"encoding/csv"
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"gorm.io/gorm"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ProjectPattern struct {
	Project string
	Total   float64
	Pattern []float64
}

func RQ3_2(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	path = path + "\\"
	patterns := AllProjectPatterns(db, projects)
	sort.Slice(patterns, func(i, j int) bool {
		c := strings.Compare(patterns[i].Project, patterns[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	/*RQ3_2_1(patterns, path+"RQ3_2_1.csv")
	RQ3_2_2(patterns, path+"RQ3_2_2.csv")
	RQ3_2_3(patterns, path+"RQ3_2_3.csv")
	RQ3_2_4(patterns, path+"RQ3_2_4.csv")*/
	fmt.Println("RQ3_2耗时:", time.Since(start).String())
	Wg.Done()
}

// RQ3_2_1 每个项目的Bug在每种模式中的数量、比例
func RQ3_2_1(patterns []ProjectPattern, path string) {

	File, _ := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	defer File.Close()
	WriterCsv := csv.NewWriter(File)
	var t [][]string

	head := []string{"Project"}
	for i := 1; i < 25; i++ {
		head = append(head, "Pattern"+strconv.Itoa(i))
		head = append(head, "Pattern"+strconv.Itoa(i)+"Percent")
	}
	t = append(t, head)

	for _, pattern := range patterns {
		row := make([]string, 0)
		row = append(row, pattern.Project)
		for _, p := range pattern.Pattern {
			row = append(row, fmt.Sprintf("%f", p))
		}
		t = append(t, row)
	}

	err := WriterCsv.WriteAll(t)
	if err != nil {
		fmt.Println("写入出错")
		return
	}
	WriterCsv.Flush() //刷新，不刷新是无法写入的
}

type ProjectPatternCount struct {
	Project      string `csv:"Project"`
	PatternCount int    `csv:"PatternCount"`
}

// RQ3_2_2 计算每个项目占了几个变更模式
func RQ3_2_2(patterns []ProjectPattern, path string) {
	ppcs := make([]ProjectPatternCount, 0)
	for _, p := range patterns {
		ppc := ProjectPatternCount{}
		ppc.Project = p.Project
		for i := 0; i < len(p.Pattern); i += 2 {
			if p.Pattern[i] != 0 {
				ppc.PatternCount++
			}
		}
		ppcs = append(ppcs, ppc)
	}
	err := csvtag.DumpToFile(ppcs, path)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// RQ3_2_3 计算每个模式有几个项目
func RQ3_2_3(patterns []ProjectPattern, path string) {
	head := make([]string, 0)
	for i := 1; i < 25; i++ {
		head = append(head, "Pattern"+strconv.Itoa(i))
	}
	t := make([][]string, 0)
	t = append(t, head)
	c := make([]int, 24)
	for _, p := range patterns {
		for i := 0; i < len(p.Pattern); i += 2 {
			if p.Pattern[i] != 0 {
				c[i/2]++
			}
		}
	}
	s := make([]string, 24)
	for i := 0; i < len(c); i++ {
		s[i] = strconv.Itoa(c[i])
	}
	t = append(t, s)

	File, _ := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	defer File.Close()
	WriterCsv := csv.NewWriter(File)
	err := WriterCsv.WriteAll(t)
	if err != nil {
		fmt.Println(err)
		return
	}
	WriterCsv.Flush()
}

// RQ3_2_4 计算每种模式的Bug数量
func RQ3_2_4(patterns []ProjectPattern, path string) {
	head := make([]string, 0)
	for i := 1; i < 25; i++ {
		head = append(head, "Pattern"+strconv.Itoa(i))
	}
	t := make([][]string, 0)
	t = append(t, head)

	c := make([]float64, 24)
	for _, p := range patterns {
		for i := 0; i < len(p.Pattern); i += 2 {
			if p.Pattern[i] != 0 {
				c[i/2] += p.Pattern[i]
			}
		}
	}
	s := make([]string, 24)
	for i := 0; i < len(c); i++ {
		s[i] = fmt.Sprintf("%.5f", c[i])
	}
	t = append(t, s)

	File, _ := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	defer File.Close()
	WriterCsv := csv.NewWriter(File)
	err := WriterCsv.WriteAll(t)
	if err != nil {
		fmt.Println(err)
		return
	}
	WriterCsv.Flush()
}

// AllProjectPatterns 获取所有项目的模式分布
func AllProjectPatterns(db *gorm.DB, projects []string) []ProjectPattern {
	pps := make([]ProjectPattern, 0)
	hadoop := ProjectPattern{}
	hadoop.Project = "Hadoop"
	hadoop.Pattern = make([]float64, 48)
	for _, project := range projects {
		pp := ProjectPatternByProject(db, project)
		if project == "HADOOP" || project == "HDFS" || project == "MAPREDUCE" || project == "YARN" {
			hadoop.Total += pp.Total
			for i := 0; i < len(pp.Pattern); i += 2 {
				hadoop.Pattern[i] += pp.Pattern[i]
			}
		} else {
			pps = append(pps, pp)
		}
	}
	for i := 0; i < len(hadoop.Pattern); i += 2 {
		hadoop.Pattern[i+1] = RoundN(100*hadoop.Pattern[i]/hadoop.Total, 1)
	}
	pps = append(pps, hadoop)
	return pps
}

// ProjectPatternByProject 获取该项目的模式分布
func ProjectPatternByProject(db *gorm.DB, project string) ProjectPattern {
	csMap := ChangeSequenceByProject(db, project)
	pp := ProjectPattern{}
	pp.Project = pro.GetNameByKey(project)
	pp.Total = float64(len(csMap))
	pp.Pattern = make([]float64, 48)
	for _, cs := range csMap {
		patternAddByCS(&pp, cs)
	}
	for i := 0; i < len(pp.Pattern); i += 2 {
		pp.Pattern[i+1] = RoundN(100*pp.Pattern[i]/pp.Total, 1)
	}
	return pp
}

func patternAddByCS(pp *ProjectPattern, cs string) {
	l := len(cs)
	//变更次数为1
	if l == 2 {
		//P1
		if cs[1] < cs[0] {
			pp.Pattern[0]++
		} else { //P2
			pp.Pattern[2]++
		}
	} else if l == 3 { //变更次数为2
		if cs[0] > cs[2] { //上升
			if cs[1] < cs[0] && cs[1] > cs[2] { //P3
				pp.Pattern[4]++
			} else if cs[1] < cs[2] { //P4
				pp.Pattern[6]++
			} else if cs[1] > cs[0] { //P5
				pp.Pattern[8]++
			}
		} else if cs[0] < cs[2] { //下降
			if cs[1] > cs[0] && cs[1] < cs[2] { //P6
				pp.Pattern[10]++
			} else if cs[1] < cs[0] { //P7
				pp.Pattern[12]++
			} else if cs[1] > cs[2] { //P8
				pp.Pattern[14]++
			}
		} else { //不变
			if cs[1] < cs[0] { //P9
				pp.Pattern[16]++
			} else { //P10
				pp.Pattern[18]++
			}
		}
	} else { //3次及以上
		o := cs[0]
		e := cs[len(cs)-1]
		if o > e { //上升
			min, max := minmax(cs)
			if cs[min] == o && cs[max] == e { //P11
				pp.Pattern[20]++
			} else if cs[min] == o && cs[max] < e { //P12
				pp.Pattern[22]++
			} else if cs[min] > o && cs[max] == e { //P13
				pp.Pattern[24]++
			} else if cs[min] > o && cs[max] < e {
				if max < min { //P14
					pp.Pattern[26]++
				} else { //P15
					pp.Pattern[28]++
				}
			}
		} else if o < e { //下降
			min, max := minmax(cs)
			if cs[max] == o && cs[min] == e { //P16
				pp.Pattern[30]++
			} else if cs[min] == e && cs[max] < o { //P17
				pp.Pattern[32]++
			} else if cs[min] > e && cs[max] == o { //P18
				pp.Pattern[34]++
			} else if cs[min] > e && cs[max] < o {
				if max < min { //P19
					pp.Pattern[36]++
				} else { //P20
					pp.Pattern[38]++
				}
			}
		} else { //不变
			min, max := minmax(cs)
			if cs[min] == o && cs[max] < o { //P21
				pp.Pattern[40]++
			} else if cs[min] > o && cs[max] == o { //P22
				pp.Pattern[42]++
			} else if cs[min] > o && cs[max] < o {
				if max < min { //P23
					pp.Pattern[44]++
				} else { //P24
					pp.Pattern[46]++
				}
			}
		}
	}
}

//返回最小值，最大值的下标
func minmax(cs string) (min, max int) {
	min = 0
	max = 0
	for i := 1; i < len(cs); i++ {
		if cs[i] < cs[max] {
			max = i
		}
		if cs[i] > cs[min] {
			min = i
		}
	}
	return
}

// ChangeSequenceByProject 获取该项目所有Bug的变更序列
func ChangeSequenceByProject(db *gorm.DB, project string) map[string]string {
	records := QueryBugRecordsByKeyPrefix(db, project)
	m := recordsToKeyRecordsMap(records)
	csMap := make(map[string]string)
	i := 1
	for k, v := range m {
		cs := CS(v)
		csMap[k] = cs
		if len(cs) >= 5 {
			fmt.Println(i, k, cs)
			i++
		}
	}
	return csMap
}

// RecordsToKeyRecordsMap 对records按key分组
func recordsToKeyRecordsMap(records []PriorityChanged) map[string][]PriorityChanged {
	m := make(map[string][]PriorityChanged)
	for _, record := range records {
		key := record.Key
		if _, ok := m[key]; !ok {
			m[key] = make([]PriorityChanged, 0)
		}
		m[key] = append(m[key], record)
	}
	return m
}

// CS 计算出某个Bug的优先级变更序列
func CS(records []PriorityChanged) string {
	cs := toId(records[0].From)
	for _, record := range records {
		cs += toId(record.To)
	}
	return cs
}

// toId 将优先级转换为对应的数字
func toId(priority string) string {
	var id string
	switch priority {
	case "Blocker":
		id = "1"
	case "Critical":
		id = "2"
	case "Major":
		id = "3"
	case "Minor":
		id = "4"
	case "Trivial":
		id = "5"
	}
	return id
}
