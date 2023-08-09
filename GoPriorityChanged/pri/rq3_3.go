package pri

import (
	"GoPriorityChanged/issue"
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

type PriorityItself struct {
	Project           string  `csv:"Project"`
	Blocker           int     `csv:"Blocker"`
	BlockerPC         int     `csv:"BlockerPC"`
	BlockerPCPercent  float64 `csv:"BlockerPCPercent"`
	Critical          int     `csv:"Critical"`
	CriticalPC        int     `csv:"CriticalPC"`
	CriticalPCPercent float64 `csv:"CriticalPCPercent"`
	Major             int     `csv:"Major"`
	MajorPC           int     `csv:"MajorPC"`
	MajorPCPercent    float64 `csv:"MajorPCPercent"`
	Minor             int     `csv:"Minor"`
	MinorPC           int     `csv:"MinorPC"`
	MinorPCPercent    float64 `csv:"MinorPCPercent"`
	Trivial           int     `csv:"Trivial"`
	TrivialPC         int     `csv:"TrivialPC"`
	TrivialPCPercent  float64 `csv:"TrivialPCPercent"`
}

func RQ3_3_1(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	priorityItselves := PriorityItselves(db, projects)
	sort.Slice(priorityItselves[:len(priorityItselves)-1], func(i, j int) bool {
		c := strings.Compare(priorityItselves[i].Project, priorityItselves[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	WritePriorityItselves(priorityItselves, path)
	fmt.Println("RQ3_3_1耗时:", time.Since(start).String())
	Wg.Done()
}

// WritePriorityItselves 写入CSV文件
func WritePriorityItselves(priorityItselves []PriorityItself, path string) {
	err := csvtag.DumpToFile(priorityItselves, path)
	if err != nil {
		fmt.Println("写入出错")
		return
	}
}

// PriorityItselves 计算所有项目的每种优先级分布及变更占比
func PriorityItselves(db *gorm.DB, projects []string) []PriorityItself {
	pis := make([]PriorityItself, 0)
	hadoop := PriorityItself{}
	hadoop.Project = "Hadoop"
	for _, project := range projects {
		priorityItself := PriorityItselfByProject(db, project)
		if project == "HADOOP" || project == "HDFS" || project == "MAPREDUCE" || project == "YARN" {
			piAccumulate(&hadoop, &priorityItself)
		} else {
			pis = append(pis, priorityItself)
		}
	}
	piPercentage(&hadoop)
	pis = append(pis, hadoop)
	pi := PriorityItself{}
	pi.Project = "All Projects"
	for _, p := range pis {
		piAccumulate(&pi, &p)
	}
	piPercentage(&pi)
	pis = append(pis, pi)
	return pis
}

// PriorityItselfByProject 计算该项目的每种优先级分布及变更占比
func PriorityItselfByProject(db *gorm.DB, project string) PriorityItself {
	issues := issue.QueryBugsByProject(db, project)
	records := QueryBugRecordsByKeyPrefix(db, project)
	m := recordsToKeyRecordsMap(records)
	pi := PriorityItself{}
	pi.Project = pro.GetNameByKey(project)
	for _, i := range issues {
		if _, ok := m[i.Key]; !ok {
			priority := i.Priority
			switch priority {
			case "Blocker":
				pi.Blocker++
			case "Critical":
				pi.Critical++
			case "Major":
				pi.Major++
			case "Minor":
				pi.Minor++
			case "Trivial":
				pi.Trivial++
			}
		} else {
			for j, pc := range m[i.Key] {
				f := pc.From
				switch f {
				case "Blocker":
					pi.Blocker++
					pi.BlockerPC++
				case "Critical":
					pi.Critical++
					pi.CriticalPC++
				case "Major":
					pi.Major++
					pi.MajorPC++
				case "Minor":
					pi.Minor++
					pi.MinorPC++
				case "Trivial":
					pi.Trivial++
					pi.TrivialPC++
					if project == "FLINK" {
						fmt.Printf("%+v\n", pc)
					}
				}
				if j == len(m[i.Key])-1 {
					t := pc.To
					switch t {
					case "Blocker":
						pi.Blocker++
					case "Critical":
						pi.Critical++
					case "Major":
						pi.Major++
					case "Minor":
						pi.Minor++
					case "Trivial":
						pi.Trivial++
					}
				}
			}
		}
	}
	piPercentage(&pi)
	return pi
}

func RQ3_3_2(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	r := make([]PriorityChanged, 0)
	for _, project := range projects {
		records := QueryBugRecordsByKeyPrefix(db, project)
		r = append(r, records...)
	}
	var m [6][6]int
	for _, record := range r {
		i, _ := strconv.Atoi(toId(record.From))
		j, _ := strconv.Atoi(toId(record.To))
		m[i-1][j-1]++
	}
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[i])-1; j++ {
			m[i][5] += m[i][j]
			if i != 5 {
				m[5][j] += m[i][j]
			}
		}
	}

	File, _ := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	defer File.Close()
	WriterCsv := csv.NewWriter(File)
	var t [][]string
	head := []string{" ", "Blocker", "Critical", "Major", "Minor", "Trivial", "All"}
	t = append(t, head)
	for i := 0; i < len(m); i++ {
		row := []string{""}
		for j := 0; j < len(m[i]); j++ {
			row = append(row, strconv.Itoa(m[i][j]))
		}
		t = append(t, row)
	}
	t[1][0] = "Blokcer"
	t[2][0] = "Critical"
	t[3][0] = "Major"
	t[4][0] = "Minor"
	t[5][0] = "Trivial"
	t[6][0] = "All"
	WriterCsv.WriteAll(t)
	WriterCsv.Flush() //刷新，不刷新是无法写入的
	fmt.Println("RQ3_3_2耗时:", time.Since(start).String())
	Wg.Done()
}

func piAccumulate(pi, p *PriorityItself) {
	pi.Blocker += p.Blocker
	pi.BlockerPC += p.BlockerPC
	pi.Critical += p.Critical
	pi.CriticalPC += p.CriticalPC
	pi.Major += p.Major
	pi.MajorPC += p.MajorPC
	pi.Minor += p.Minor
	pi.MinorPC += p.MinorPC
	pi.Trivial += p.Trivial
	pi.TrivialPC += p.TrivialPC
}

func piPercentage(pi *PriorityItself) {
	pi.BlockerPCPercent = RoundN(100*float64(pi.BlockerPC)/float64(pi.Blocker), 1)
	pi.CriticalPCPercent = RoundN(100*float64(pi.CriticalPC)/float64(pi.Critical), 1)
	pi.MajorPCPercent = RoundN(100*float64(pi.MajorPC)/float64(pi.Major), 1)
	pi.MinorPCPercent = RoundN(100*float64(pi.MinorPC)/float64(pi.Minor), 1)
	pi.TrivialPCPercent = RoundN(100*float64(pi.TrivialPC)/float64(pi.Trivial), 1)
}
