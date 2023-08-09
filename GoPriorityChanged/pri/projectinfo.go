package pri

import (
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"math"
	"sort"
	"strings"
	"time"
)

type ProjectInfo struct {
	Project      string `csv:"Project"`
	Age          int    `csv:"Age(y)"`
	Language     string `csv:"Language"`
	Star         int    `csv:"#Star"`
	Contributors int    `csv:"#Contributors"`
	Revision     int    `csv:"#Revision"`
	Committers   int    `csv:"#Committers"`
	Bug          int    `csv:"#Bug in JIRA"`
}

func GetProjectInfo(commitFilePath string, outDir string) {
	pis := make([]ProjectInfo, 0)
	m := GetProComFileMap(commitFilePath)
	for name, paths := range m { //循环每个项目
		pi := ProjectInfoByProject(name, paths)
		pis = append(pis, pi)
	}
	sort.Slice(pis, func(i, j int) bool {
		c := strings.Compare(pis[i].Project, pis[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	err := csvtag.DumpToFile(pis, outDir+"\\PI.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ProjectInfoByProject(name string, paths []string) ProjectInfo {
	pi := ProjectInfo{}
	pi.Project = name
	commits := make([]Commit, 0)
	for _, v := range paths { //循环每个commit文件
		lines := ReadCommitFile(v)
		commits = append(commits, ParseCommit(lines)...)
	}
	pi.Revision = len(commits)
	age, c := getAgeAndCommitters(commits)
	pi.Age = age
	pi.Committers = c
	return pi
}

//获取项目的年龄和提交者数量
func getAgeAndCommitters(commits []Commit) (age, committers int) {
	var h float64 = 0
	c := make(map[string]bool)
	for _, commit := range commits {
		hours := time.Since(commit.Date).Hours()
		if hours > h {
			h = hours
		}
		c[commit.Author] = true
	}
	age = int(math.Floor(h/(24*365) + 0.5))
	committers = len(c)
	return
}
