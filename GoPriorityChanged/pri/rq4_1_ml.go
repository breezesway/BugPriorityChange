package pri

import (
	"GoPriorityChanged/pro"
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"gorm.io/gorm"
	"math"
	"sort"
	"strconv"
	"strings"
)

type ProjectMLBugSLBug struct {
	Name                 string  `csv:"Name"`
	Key                  string  `csv:"Key"`
	MLBugCount           int     `csv:"MLBugCount"`
	SLBugCount           int     `csv:"SLBugCount"`
	MLIncCount           int     `csv:"MLIncCount"`
	MLIncPercent         float64 `csv:"MLIncPercent"`
	SLIncCount           int     `csv:"SLIncCount"`
	SLIncPercent         float64 `csv:"SLIncPercent"`
	MLDecCount           int     `csv:"MLDecCount"`
	MLDecPercent         float64 `csv:"MLDecPercent"`
	SLDecCount           int     `csv:"SLDecCount"`
	SLDecPercent         float64 `csv:"SLDecPercent"`
	MLRestorationCount   int     `csv:"MLRestorationCount"`
	MLRestorationPercent float64 `csv:"MLRestorationPercent"`
	SLRestorationCount   int     `csv:"SLRestorationCount"`
	SLRestorationPercent float64 `csv:"SLRestorationPercent"`
	MLChangeCountAve     float64 `csv:"MLChangeCountAve"`
	SLChangeCountAve     float64 `csv:"SLChangeCountAve"`
	MLChangeRangeAve     float64 `csv:"MLChangeRangeAve"`
	SLChangeRangeAve     float64 `csv:"SLChangeRangeAve"`
}

type MLBugSLBug struct {
	Key   string
	IsML  bool
	Count int
	Range float64
	Trend string
}

func RQ4_1_ML(db *gorm.DB, path string, out string) {
	pmses := make([]ProjectMLBugSLBug, 0)
	m := GetProComFileMap(path)
	pcMap := QueryAllIssue(db)
	for name, paths := range m { //循环遍历每个name
		keys, err := pro.GetKeyByName(name)
		if err != nil {
			fmt.Println(name, err)
		}
		for _, key := range keys { //循环每个key
			index := containKey2(pmses, key)
			var pi *ProjectMLBugSLBug
			if index == -1 {
				pi = &ProjectMLBugSLBug{}
				pi.Name = name
				pi.Key = key
				pmses = append(pmses, *pi)
			}
			index = containKey2(pmses, key)
			pi = &pmses[index]
			mss := make([]MLBugSLBug, 0)
			for _, v := range paths { //循环每个commit文件
				lines := ReadCommitFile(v)
				commits := ParseCommit(lines)
				bcMap := bugCommitMap(db, key, commits)
				for k, v := range bcMap {
					if _, ok := pcMap[k]; ok {
						ms := calMLBugSLBug(k, v, commits)
						mss = append(mss, ms)
					}
				}
			}
			pms := calProjectMLBugSLBug(mss)
			pms.Name = name
			pms.Key = key
			accumulateProjectMLBugSLBug(pi, pms)
		}
	}
	hadoop := ProjectMLBugSLBug{}
	for i := 0; i < len(pmses); i++ {
		key := pmses[i].Key
		if key == "HADOOP" || key == "HDFS" || key == "MAPREDUCE" || key == "YARN" {
			accumulateProjectMLBugSLBug(&hadoop, pmses[i])
			fmt.Println(hadoop.MLChangeCountAve, hadoop.SLChangeCountAve, hadoop.MLChangeRangeAve, hadoop.SLChangeRangeAve)
			pmses = append(pmses[:i], pmses[i+1:]...)
			i--
		}
	}
	hadoop.Name = "hadoop"
	hadoop.Key = "HADOOP, MAPREDUCE, YARN, HDFS"
	pmses = append(pmses, hadoop)

	fmt.Println("________")

	all := ProjectMLBugSLBug{}
	for i := 0; i < len(pmses); i++ {
		accumulateProjectMLBugSLBug(&all, pmses[i])
		fmt.Println(all.MLChangeCountAve, all.SLChangeCountAve, all.MLChangeRangeAve, all.SLChangeRangeAve)
	}
	all.Name = "all projects"
	all.Key = "All"
	pmses = append(pmses, all)
	sort.Slice(pmses, func(i, j int) bool {
		c := strings.Compare(pmses[i].Name, pmses[j].Name)
		if c < 0 {
			return true
		} else {
			return false
		}
	})

	err := csvtag.DumpToFile(pmses, out)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func calMLBugSLBug(key string, indexs []int, commits []Commit) MLBugSLBug {
	ms := MLBugSLBug{}
	ms.Key = key
	lMap := make(map[string]bool)
	for _, index := range indexs {
		for _, file := range commits[index].Files {
			lMap[SufMap[file.Suf]] = true
		}
	}
	if len(lMap) == 1 {
		ms.IsML = false
	} else {
		ms.IsML = true
	}
	pcs := RecordsMap[key]
	ms.Count = len(pcs)
	var pcRange float64 = 0
	for _, pc := range pcs {
		to, _ := strconv.Atoi(toId(pc.To))
		from, _ := strconv.Atoi(toId(pc.From))
		pcRange += math.Abs(float64(to) - float64(from))
	}
	ms.Range = pcRange / float64(len(pcs))
	to, _ := strconv.Atoi(toId(pcs[len(pcs)-1].To))
	from, _ := strconv.Atoi(toId(pcs[0].From))
	if to < from {
		ms.Trend = "Inc"
	} else if to > from {
		ms.Trend = "Dec"
	} else {
		ms.Trend = "Restoration"
	}
	return ms
}

func calProjectMLBugSLBug(mss []MLBugSLBug) ProjectMLBugSLBug {
	pms := ProjectMLBugSLBug{}
	for _, ms := range mss {
		if ms.IsML {
			pms.MLBugCount++
			if ms.Trend == "Inc" {
				pms.MLIncCount++
			} else if ms.Trend == "Dec" {
				pms.MLDecCount++
			} else if ms.Trend == "Restoration" {
				pms.MLRestorationCount++
			}
			pms.MLChangeCountAve += float64(ms.Count)
			pms.MLChangeRangeAve += ms.Range
		} else {
			pms.SLBugCount++
			if ms.Trend == "Inc" {
				pms.SLIncCount++
			} else if ms.Trend == "Dec" {
				pms.SLDecCount++
			} else if ms.Trend == "Restoration" {
				pms.SLRestorationCount++
			}
			pms.SLChangeCountAve += float64(ms.Count)
			pms.SLChangeRangeAve += ms.Range
		}
	}
	t1 := float64(pms.MLBugCount)
	t2 := float64(pms.SLBugCount)
	if t1 != 0 {
		pms.MLChangeCountAve /= t1
		pms.MLChangeRangeAve /= t1
		pms.MLIncPercent = float64(pms.MLIncCount) / t1
		pms.MLDecPercent = float64(pms.MLDecCount) / t1
		pms.MLRestorationPercent = float64(pms.MLRestorationCount) / t1
	}
	if t2 != 0 {
		pms.SLChangeCountAve /= t2
		pms.SLChangeRangeAve /= t2
		pms.SLIncPercent = float64(pms.SLIncCount) / t2
		pms.SLDecPercent = float64(pms.SLDecCount) / t2
		pms.SLRestorationPercent = float64(pms.SLRestorationCount) / t2
	}
	return pms
}

func accumulateProjectMLBugSLBug(pi *ProjectMLBugSLBug, pms ProjectMLBugSLBug) {
	mlTotal := float64(pi.MLBugCount + pms.MLBugCount)
	slTotal := float64(pi.SLBugCount + pms.SLBugCount)
	pi.MLIncCount += pms.MLIncCount
	pi.MLIncPercent = float64(pi.MLIncCount) / mlTotal
	pi.MLDecCount += pms.MLDecCount
	pi.MLDecPercent = float64(pi.MLDecCount) / mlTotal
	pi.MLRestorationCount += pms.MLRestorationCount
	pi.MLRestorationPercent = float64(pi.MLRestorationCount) / mlTotal
	pi.SLIncCount += pms.SLIncCount
	pi.SLIncPercent = float64(pi.SLIncCount) / slTotal
	pi.SLDecCount += pms.SLDecCount
	pi.SLDecPercent = float64(pi.SLDecCount) / slTotal
	pi.SLRestorationCount += pms.SLRestorationCount
	pi.SLRestorationPercent = float64(pi.SLRestorationCount) / slTotal
	pi.MLChangeCountAve = (pi.MLChangeCountAve*float64(pi.MLBugCount) + pms.MLChangeCountAve*float64(pms.MLBugCount)) / mlTotal
	pi.SLChangeCountAve = (pi.SLChangeCountAve*float64(pi.SLBugCount) + pms.SLChangeCountAve*float64(pms.SLBugCount)) / slTotal
	pi.MLChangeRangeAve = (pi.MLChangeRangeAve*float64(pi.MLBugCount) + pms.MLChangeRangeAve*float64(pms.MLBugCount)) / mlTotal
	pi.SLChangeRangeAve = (pi.SLChangeRangeAve*float64(pi.SLBugCount) + pms.SLChangeRangeAve*float64(pms.SLBugCount)) / slTotal
	pi.MLBugCount += pms.MLBugCount
	pi.SLBugCount += pms.SLBugCount
}
