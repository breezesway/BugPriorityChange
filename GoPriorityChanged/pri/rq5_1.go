package pri

import (
	"GoPriorityChanged/comment"
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

type PersonCount struct {
	Project                 string  `csv:"Project"`
	Creator                 int     `csv:"Creator"`
	CreatorP                float64 `csv:"CreatorP"`
	Assignee                int     `csv:"Assignee"`
	AssigneeP               float64 `csv:"AssigneeP"`
	CreatorAssignee         int     `csv:"CreatorAssignee"`
	CreatorAssigneeP        float64 `csv:"CreatorAssigneeP"`
	Commenter               int     `csv:"Commenter"`
	CommenterP              float64 `csv:"CommenterP"`
	FieldModifier           int     `csv:"FieldModifier"`
	FieldModifierP          float64 `csv:"FieldModifierP"`
	CommenterFieldModifier  int     `csv:"CommenterFieldModifier"`
	CommenterFieldModifierP float64 `csv:"CommenterFieldModifierP"`
	Others                  int     `csv:"Others"`
	OthersP                 float64 `csv:"OthersP"`
	Total                   int     `csv:"Total"`
}

func RQ5_1(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	idpcs := PersonCounts(db, projects, false)
	idpcs2 := PersonCounts(db, projects, true)
	sort.Slice(idpcs[:len(idpcs)-1], func(i, j int) bool {
		c := strings.Compare(idpcs[i].Project, idpcs[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	sort.Slice(idpcs2[:len(idpcs2)-1], func(i, j int) bool {
		c := strings.Compare(idpcs2[i].Project, idpcs2[j].Project)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	err := csvtag.DumpToFile(idpcs, path+"\\RQ5_1.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = csvtag.DumpToFile(idpcs2, path+"\\RQ5_1(duplicated).csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("RQ5_1耗时：", time.Since(start).String())
	Wg.Done()
}

// PersonCounts 计算所有项目三种人员的人数分布情况
func PersonCounts(db *gorm.DB, projects []string, isDuplicated bool) []PersonCount {
	idpcs := make([]PersonCount, 0)
	hadoop := PersonCount{}
	hadoop.Project = "Hadoop"
	for _, project := range projects {
		idpc := PersonCountByProject(db, project, isDuplicated)
		if idpc.Project == "Hadoop" {
			calPCAccumulate(&hadoop, &idpc)
		} else {
			idpcs = append(idpcs, idpc)
		}
	}
	setPercent(&hadoop)
	idpc := PersonCount{}
	idpc.Project = "All"
	for _, i := range idpcs {
		calPCAccumulate(&idpc, &i)
	}
	setPercent(&idpc)
	idpcs = append(idpcs, hadoop)
	idpcs = append(idpcs, idpc)
	return idpcs
}

// PersonCountByProject 计算该项目三种人员的人数分布情况
func PersonCountByProject(db *gorm.DB, project string, isDuplicated bool) PersonCount {
	records := QueryBugRecordsByKeyPrefix(db, project)
	idpc := PersonCount{}
	idpc.Project = pro.GetNameByKey(project)
	idpc.Total = len(records)
	for _, record := range records {
		key := record.Key
		iss := issue.QueryIssueByKey(db, key)
		judgeTypeAndAdd(db, &idpc, record.Author, iss, isDuplicated)
	}
	setPercent(&idpc)
	return idpc
}

//判断这个人属于哪种类型，并在相应类型上加一
func judgeTypeAndAdd(db *gorm.DB, idpc *PersonCount, author string, iss issue.Issue, isDuplicated bool) {
	if isDuplicated {
		duplicateStatistics(db, idpc, author, iss)
	} else {
		noDuplicateStatistics(db, idpc, author, iss)
	}
}

//重复统计
func duplicateStatistics(db *gorm.DB, idpc *PersonCount, author string, iss issue.Issue) {
	creator := iss.Creator
	assignee := iss.Assignee
	if author == creator {
		idpc.Creator++
	}
	if author == assignee {
		idpc.Assignee++
	}
	commentersMap := comment.QueryCommentersByRowIds(db, iss.Comments)
	modifiersMap := his.QueryFieldModifiersByRowIds(db, iss.HistoryIds)
	_, commenterOk := commentersMap[author]
	_, modifierOk := modifiersMap[author]
	if commenterOk {
		idpc.Commenter++
	}
	if modifierOk {
		idpc.FieldModifier++
	}
	if author != creator && author != assignee && !commenterOk && !modifierOk {
		idpc.Others++
	}

}

//非重复统计
func noDuplicateStatistics(db *gorm.DB, idpc *PersonCount, author string, iss issue.Issue) {
	creator := iss.Creator
	assignee := iss.Assignee
	if author == creator && author != assignee {
		idpc.Creator++
	} else if author != creator && author == assignee {
		idpc.Assignee++
	} else if author == creator && author == assignee {
		idpc.CreatorAssignee++
	} else {
		commentersMap := comment.QueryCommentersByRowIds(db, iss.Comments)
		modifiersMap := his.QueryFieldModifiersByRowIds(db, iss.HistoryIds)
		_, commenterOk := commentersMap[author]
		_, modifierOk := modifiersMap[author]
		if commenterOk && !modifierOk {
			idpc.Commenter++
		} else if !commenterOk && modifierOk {
			idpc.FieldModifier++
		} else if commenterOk && modifierOk {
			idpc.CommenterFieldModifier++
		} else {
			idpc.Others++
		}
	}
}

func calPCAccumulate(idpc, i *PersonCount) {
	idpc.Creator += i.Creator
	idpc.Assignee += i.Assignee
	idpc.CreatorAssignee += i.CreatorAssignee
	idpc.Commenter += i.Commenter
	idpc.FieldModifier += i.FieldModifier
	idpc.CommenterFieldModifier += i.CommenterFieldModifier
	idpc.Others += i.Others
	idpc.Total += i.Total
}

//计算下各个百分比
func setPercent(idpc *PersonCount) {
	t := float64(idpc.Total)
	idpc.CreatorP = RoundN(100*float64(idpc.Creator)/t, 1)
	idpc.AssigneeP = RoundN(100*float64(idpc.Assignee)/t, 1)
	idpc.CreatorAssigneeP = RoundN(100*float64(idpc.CreatorAssignee)/t, 1)
	idpc.CommenterP = RoundN(100*float64(idpc.Commenter)/t, 1)
	idpc.FieldModifierP = RoundN(100*float64(idpc.FieldModifier)/t, 1)
	idpc.CommenterFieldModifierP = RoundN(100*float64(idpc.CommenterFieldModifier)/t, 1)
	idpc.OthersP = RoundN(100*float64(idpc.Others)/t, 1)
}
