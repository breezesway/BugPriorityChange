package pri

import (
	"GoPriorityChanged/comment"
	"GoPriorityChanged/issue"
	"GoPriorityChanged/pro"
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"gorm.io/gorm"
	"sort"
	"strings"
	"time"
)

type ProjectCommentComplexity struct {
	Project string  `csv:"Project"`
	CCAveC  float64 `csv:"CommentCountAveC"`
	CCAveN  float64 `csv:"CommentCountAveN"`
	CLAveC  float64 `csv:"CommentLengthAveC"`
	CLAveN  float64 `csv:"CommentLengthAveN"`
	CPCAveC float64 `csv:"CommentPeopleCountAveC"`
	CPCAveN float64 `csv:"CommentPeopleCountAveN"`
}

type BugCommentC struct {
	Key          string `csv:"Key"`
	CCount       int    `csv:"CommentCount"`
	CLength      int    `csv:"CommentLength"`
	CPeopleCount int    `csv:"CommentPeopleCount"`
}

func RQ4_2(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	pcscs := CommentComplexities(db, projects)
	sort.Slice(pcscs, func(i, j int) bool {
		c := strings.Compare(pcscs[i].Project, pcscs[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	err := csvtag.DumpToFile(pcscs, path)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("RQ4_2耗时:", time.Since(start).String())
	Wg.Done()
}

func CommentComplexities(db *gorm.DB, projects []string) []ProjectCommentComplexity {
	pcscs := make([]ProjectCommentComplexity, 0)
	pcMap := QueryAllIssue(db)
	hadoop := ProjectCommentComplexity{}
	hadoop.Project = "Hadoop"
	hadoopBugCC := make([]BugCommentC, 0)
	hadoopBugCN := make([]BugCommentC, 0)
	for _, project := range projects {
		bugCC, bugCN := BugCBugNCommentsByProject(pcMap, db, project)
		if project == "HADOOP" || project == "HDFS" || project == "MAPREDUCE" || project == "YARN" {
			hadoopBugCC = append(hadoopBugCC, bugCC...)
			hadoopBugCN = append(hadoopBugCN, bugCN...)
		} else {
			pcsc := ProjectCommentComplexity{}
			pcsc.Project = pro.GetNameByKey(project)
			calCommentSpendAve(&pcsc, bugCC, bugCN)
			csvtag.DumpToFile(bugCC, "C:\\Users\\MasterCai\\Desktop\\4_2\\"+project+"C.csv")
			csvtag.DumpToFile(bugCN, "C:\\Users\\MasterCai\\Desktop\\4_2\\"+project+"N.csv")
			pcscs = append(pcscs, pcsc)
		}
	}
	csvtag.DumpToFile(hadoopBugCC, "C:\\Users\\MasterCai\\Desktop\\4_2\\hadoopC.csv")
	csvtag.DumpToFile(hadoopBugCN, "C:\\Users\\MasterCai\\Desktop\\4_2\\hadoopN.csv")
	calCommentSpendAve(&hadoop, hadoopBugCC, hadoopBugCN)
	pcscs = append(pcscs, hadoop)
	return pcscs
}

// BugCBugNCommentsByProject 按优先级改变和没改变的两列返回
func BugCBugNCommentsByProject(pcMap map[string]bool, db *gorm.DB, project string) (BugC, BugN []BugCommentC) {
	BugC = make([]BugCommentC, 0)
	BugN = make([]BugCommentC, 0)
	issues := issue.QueryBugsByProject(db, project)
	for _, iss := range issues {
		bcc := getBugCommentC(db, iss)
		if _, ok := pcMap[bcc.Key]; ok {
			BugC = append(BugC, bcc)
		} else {
			BugN = append(BugN, bcc)
		}
	}
	return
}

//获取该Bug的Comment相关的数据
func getBugCommentC(db *gorm.DB, issue issue.Issue) BugCommentC {
	bcc := BugCommentC{}
	bcc.Key = issue.Key
	s := issue.Comments[1 : len(issue.Comments)-1]
	commentIds := strings.Split(s, ", ")
	authorMap := make(map[string]bool)
	for _, id := range commentIds {
		c := comment.QueryCommentById(db, id)
		bcc.CLength += len(c.Body)
		authorMap[c.Author] = true
	}
	bcc.CCount = len(commentIds)
	bcc.CPeopleCount = len(authorMap)
	return bcc
}

//计算平均值
func calCommentSpendAve(pcsc *ProjectCommentComplexity, bugCC []BugCommentC, bugCN []BugCommentC) {
	for _, c := range bugCC {
		pcsc.CCAveC += float64(c.CCount)
		pcsc.CLAveC += float64(c.CLength)
		pcsc.CPCAveC += float64(c.CPeopleCount)
	}
	l := float64(len(bugCC))
	pcsc.CCAveC = RoundN(pcsc.CCAveC/l, 2)
	pcsc.CLAveC = RoundN(pcsc.CLAveC/l, 0)
	pcsc.CPCAveC = RoundN(pcsc.CPCAveC/l, 2)

	for _, c := range bugCN {
		pcsc.CCAveN += float64(c.CCount)
		pcsc.CLAveN += float64(c.CLength)
		pcsc.CPCAveN += float64(c.CPeopleCount)
	}
	l = float64(len(bugCN))
	pcsc.CCAveN = RoundN(pcsc.CCAveN/l, 2)
	pcsc.CLAveN = RoundN(pcsc.CLAveN/l, 0)
	pcsc.CPCAveN = RoundN(pcsc.CPCAveN/l, 2)
}
