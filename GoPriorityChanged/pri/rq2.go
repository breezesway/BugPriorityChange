package pri

import (
	"GoPriorityChanged/his"
	"GoPriorityChanged/issue"
	"GoPriorityChanged/pro"
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"gorm.io/gorm"
	"sort"
	"strings"
	"time"
)

type Interval struct {
	Project        string  `csv:"Project"`
	Min            float64 `csv:"Min"`
	Median         float64 `csv:"Median"`
	Average        float64 `csv:"Average"`
	Max            float64 `csv:"Max"`
	Over365Count   int     `csv:"Over365Count"`
	Over365Percent float64 `csv:"Over365Percent"`
}

func RQ2_1(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	m := TimeIntervals(db, projects)
	/*for k, v := range m {
		File, _ := os.OpenFile("C:\\Users\\MasterCai\\Desktop\\2_1\\"+k+".csv", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)

		WriterCsv := csv.NewWriter(File)
		var t [][]string

		head := []string{k}
		t = append(t, head)

		for _, interval := range v {
			row := make([]string, 0)
			row = append(row, fmt.Sprintf("%f", interval))
			t = append(t, row)
		}
		WriterCsv.WriteAll(t)
		WriterCsv.Flush() //刷新，不刷新是无法写入的
		File.Close()
	}*/
	intervals := Metric(m)
	sort.Slice(intervals, func(i, j int) bool {
		c := strings.Compare(intervals[i].Project, intervals[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	WriteTimeIntervals(path, intervals)
	fmt.Println("RQ2_1耗时：", time.Since(start).String())
	Wg.Done()
}

// ChangeTimeInterval 计算该Bug从创建到第一次修改优先级的间隔时间
func ChangeTimeInterval(db *gorm.DB, record PriorityChanged) float64 {
	var created string
	db.Table("issue").Select("created").Where("issue.key = ?", record.Key).Find(&created)
	pc := record.Created
	d := timeSub(pc, created)
	//strconv.ParseFloat(fmt.Sprintf("%.4f", d.Hours()/24), 64)
	hour := d.Hours() / 24
	return hour
}

// TimeIntervalByProject 计算该项目所有Bug的TimeInterval
func TimeIntervalByProject(db *gorm.DB, project string) []float64 {
	records := QueryBugRecordsByKeyPrefix(db, project)
	m := make(map[string]float64)
	intervals := make([]float64, 0)
	for _, record := range records {
		if _, ok := m[record.Key]; !ok {
			interval := ChangeTimeInterval(db, record)
			m[record.Key] = interval
			intervals = append(intervals, interval)
		}
	}
	return intervals
}

// TimeIntervals 计算每个项目所有Bug的TimeInterval
func TimeIntervals(db *gorm.DB, projects []string) map[string][]float64 {
	m := make(map[string][]float64)
	hadoop := make([]float64, 0)
	for _, project := range projects {
		vals := TimeIntervalByProject(db, project)
		if project == "HADOOP" || project == "HDFS" || project == "MAPREDUCE" || project == "YARN" {
			hadoop = append(hadoop, vals...)
		} else {
			m[project] = vals
		}
	}
	m["HADOOP"] = hadoop
	return m
}

// Metric 根据给定的每个Bug的TimeInterval，计算每个项目的各项指标
func Metric(m map[string][]float64) []Interval {
	intervals := make([]Interval, 0)
	for k, v := range m {
		sort.Float64s(v)
		l := len(v)
		var min = v[0]
		var max = v[l-1]
		var median float64
		if l%2 == 0 {
			median = (v[l/2-1] + v[l/2]) / 2
		} else {
			median = v[l/2]
		}
		var average float64 = 0
		var over365 int = 0
		for _, day := range v {
			average += day
			if day > 365 {
				over365++
			}
		}
		average = average / float64(l)
		interval := Interval{
			Project:        pro.GetNameByKey(k),
			Min:            min,
			Median:         RoundN(median, 2),
			Average:        RoundN(average, 2),
			Max:            RoundN(max, 2),
			Over365Count:   over365,
			Over365Percent: RoundN(100*float64(over365)/float64(l), 1),
		}
		intervals = append(intervals, interval)
	}
	return intervals
}

// WriteTimeIntervals 将每个项目的各项指标写入csv文件
func WriteTimeIntervals(path string, intervals []Interval) {
	err := csvtag.DumpToFile(intervals, path)
	if err != nil {
		return
	}
}

type ChangePhase struct {
	Project       string  `csv:"Project"`
	Total         int     `csv:"Total"`
	PhaseBCount   int     `csv:"PhaseBCount"`
	PhaseBPercent float64 `csv:"PhaseBPercent"`
	PhaseICount   int     `csv:"PhaseICount"`
	PhaseIPercent float64 `csv:"PhaseIPercent"`
	PhaseRCount   int     `csv:"PhaseRCount"`
	PhaseRPercent float64 `csv:"PhaseRPercent"`
	PhaseACount   int     `csv:"PhaseACount"`
	PhaseAPercent float64 `csv:"PhaseAPercent"`
}

func RQ2_2(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	changePhases := ChangePhases(db, projects)
	sort.Slice(changePhases[:len(changePhases)-1], func(i, j int) bool {
		c := strings.Compare(changePhases[i].Project, changePhases[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	WriteChangePhases(path, changePhases)
	fmt.Println("RQ2_2耗时:", time.Since(start).String())
	Wg.Done()
}

// ChangePhases 获取所有项目的ChangePhase分布情况
func ChangePhases(db *gorm.DB, projects []string) []ChangePhase {
	changePhases := make([]ChangePhase, 0)
	hadoop := ChangePhase{}
	hadoop.Project = "Hadoop"
	for _, project := range projects {
		changePhase := ChangePhaseByProject(db, project)
		if project == "HADOOP" || project == "HDFS" || project == "MAPREDUCE" || project == "YARN" {
			cpAccumulate(&hadoop, &changePhase)
		} else {
			changePhases = append(changePhases, changePhase)
		}
	}
	cpPercentage(&hadoop)
	changePhases = append(changePhases, hadoop)
	c := ChangePhase{}
	c.Project = "All projects"
	for _, p := range changePhases {
		cpAccumulate(&c, &p)
	}
	cpPercentage(&c)
	changePhases = append(changePhases, c)
	return changePhases
}

// WriteChangePhases 将所有项目的ChangePhase写入CSV文件
func WriteChangePhases(path string, changePhases []ChangePhase) {
	err := csvtag.DumpToFile(changePhases, path)
	if err != nil {
		return
	}
}

// ChangePhaseByProject 计算该项目的ChangePhase分布情况
func ChangePhaseByProject(db *gorm.DB, project string) ChangePhase {
	records := QueryBugRecordsByKeyPrefix(db, project)
	c := ChangePhase{}
	c.Project = pro.GetNameByKey(project)
	c.Total = len(records)
	for _, record := range records {
		issueRecord := issue.QueryIssueByKey(db, record.Key)
		histories := his.QueryHistoriesByRowIds(db, issueRecord.HistoryIds)
		phase := judgePhase(record, histories)
		switch phase {
		case "B":
			c.PhaseBCount++
		case "I":
			c.PhaseICount++
		case "R":
			c.PhaseRCount++
		case "A":
			c.PhaseACount++
		}
	}
	cpPercentage(&c)
	return c
}

//累加
func cpAccumulate(c, p *ChangePhase) {
	c.Total += p.Total
	c.PhaseBCount += p.PhaseBCount
	c.PhaseICount += p.PhaseICount
	c.PhaseRCount += p.PhaseRCount
	c.PhaseACount += p.PhaseACount
}

func cpPercentage(c *ChangePhase) {
	f := float64(c.Total)
	c.PhaseBPercent = RoundN(100*float64(c.PhaseBCount)/f, 1)
	c.PhaseIPercent = RoundN(100*float64(c.PhaseICount)/f, 1)
	c.PhaseRPercent = RoundN(100*float64(c.PhaseRCount)/f, 1)
	c.PhaseAPercent = RoundN(100*float64(c.PhaseACount)/f, 1)
}

//判断该Bug属于哪个Phase
func judgePhase(record PriorityChanged, histories []his.History) string {
	var phase string
	for i := 0; i < len(histories); i++ {
		hisCreated := histories[i].His.Created
		items := histories[i].Items
		for j := 0; j < len(items); j++ {
			if items[j].Field == "status" {
				//S1
				if items[j].ToString == "Closed" || items[j].ToString == "Resolved" {
					if timeSub(record.Created, hisCreated).Seconds() > 0 {
						phase = "A"
					}
				}
				//S2
				if items[j].ToString == "Reopened" {
					if timeSub(record.Created, hisCreated).Seconds() > 0 {
						phase = "R"
					}
				}
			}
		}
	}
	if len(phase) != 0 {
		return phase
	}
	for i := 0; i < len(histories); i++ {
		hisCreated := histories[i].His.Created
		items := histories[i].Items
		for j := 0; j < len(items); j++ {
			//S3
			if items[j].Field == "status" && items[j].ToString == "In Progress" || items[j].Field == "Patch Info" && strings.EqualFold(items[j].ToString, "[Patch available]") {
				if timeSub(record.Created, hisCreated).Seconds() > 0 { //S3.1
					return "I"
				} else { //S3.2
					return "B"
				}
			}
			//S4
			if items[j].Field == "Link" || items[j].Field == "RemoteIssueLink" {
				if timeSub(record.Created, hisCreated).Seconds() > 0 { //S4.1
					phase = "I"
				} else { //S4.2
					phase = "B"
				}
			}
		}
	}
	if len(phase) != 0 {
		return phase
	}
	//优先级改变是第一条历史记录
	for _, item := range histories[0].Items { //S5.1
		if item.Field == "priority" {
			return "B"
		}
	}
	//优先级改变不是第一条历史记录
	return "I" ////S5.2
}

func timeSub(s1 string, s2 string) time.Duration {
	t1, _ := time.Parse("2006-01-02T15:04:05.000+0000", s1)
	t2, _ := time.Parse("2006-01-02T15:04:05.000+0000", s2)
	d := t1.Sub(t2)
	return d
}

func StringToDateTime(s1 string) time.Time {
	t1, _ := time.Parse("2006-01-02T15:04:05.000+0000", s1)
	return t1
}
