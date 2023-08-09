package pri

import (
	"GoPriorityChanged/comment"
	"GoPriorityChanged/his"
	"GoPriorityChanged/issue"
	"GoPriorityChanged/pro"
	"fmt"
	"gorm.io/gorm"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ProjectPerson struct {
	Name            string  `csv:"Name"`
	CreateTotal     int     `csv:"CreateTotal"`
	CreatePreferC   int     `csv:"CreatePreferC"`
	CreatePreferP   float64 `csv:"CreatePreferP"`
	AssigneeTotal   int     `csv:"AssigneeTotal"`
	AssigneePreferC int     `csv:"AssigneePreferC"`
	AssigneePreferP float64 `csv:"AssigneePreferP"`
	ModifyTotal     int     `csv:"ModifyTotal"`
	ModifyPreferC   int     `csv:"ModifyPreferC"`
	ModifyPreferP   float64 `csv:"ModifyPreferP"`
	IncPreferC      int     `csv:"IncPreferC"`
	IncPreferP      float64 `csv:"IncPreferP"`
	DecPreferC      int     `csv:"DecPreferC"`
	DecPreferP      float64 `csv:"DecPreferP"`
	RePreferC       int     `csv:"RePreferC"`
	RePreferP       float64 `csv:"RePreferP"`
}

type PersonOperation struct {
	Name            string  `csv:"Name"`
	CreateCount     int     `csv:"CreateCount"`
	CreateCountC    int     `csv:"CreateCountC"`
	CreatePercent   float64 `csv:"CreatePercent"`
	AssigneeCount   int     `csv:"AssigneeCount"`
	AssigneeCountC  int     `csv:"AssigneeCountC"`
	AssigneePercent float64 `csv:"AssigneePercent"`
	Involvement     int     `csv:"Involvement"`
	ModifyCount     int     `csv:"ModifyCount"`
	ModifyPercent   float64 `csv:"ModifyPercent"`
	IncModifyCount  int     `csv:"IncModifyCount"`
	DecModifyCount  int     `csv:"DecModifyCount"`
}

func RQ5_2(db *gorm.DB, projects []string, path string) {
	start := time.Now()
	pps := ProjectPersons(db, projects)
	//fmt.Println(totalCre/lenTotal, totalAss/lenTotal, totalInv/lenTotal)
	sort.Slice(pps, func(i, j int) bool {
		c := strings.Compare(pps[i].Name, pps[j].Name)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	/*err := csvtag.DumpToFile(pps, path)
	if err != nil {
		fmt.Println(err)
		return
	}*/
	fmt.Println("RQ5_2耗时：", time.Since(start).String())
	Wg.Done()
}

var totalCre = 0
var totalAss = 0
var totalInv = 0
var lenTotal = 0
var bcMap = make(map[string]float64)

func ProjectPersons(db *gorm.DB, projects []string) []ProjectPerson {
	pps := make([]ProjectPerson, 0)
	hadoop := ProjectPerson{}
	hadoop.Name = "Hadoop"
	posHadoop := make([]PersonOperation, 0)
	for _, project := range projects {
		pp, pos := ProjectPersonByProject(db, project)
		if pp.Name == "Hadoop" {
			hadoop.CreateTotal += pp.CreateTotal
			hadoop.CreatePreferC += pp.CreatePreferC
			hadoop.AssigneeTotal += pp.AssigneeTotal
			hadoop.AssigneePreferC += pp.AssigneePreferC
			hadoop.ModifyTotal += pp.ModifyTotal
			hadoop.ModifyPreferC += pp.ModifyPreferC
			hadoop.IncPreferC += pp.IncPreferC
			hadoop.DecPreferC += pp.DecPreferC
			hadoop.RePreferC += pp.RePreferC
			posHadoop = append(posHadoop, pos...)
		} else {
			pps = append(pps, pp)
			//csvtag.DumpToFile(pos, "C:\\Users\\MasterCai\\Desktop\\5_2\\"+project+".csv")
		}
	}
	calProjectPercent(&hadoop)
	pps = append(pps, hadoop)
	sort.Slice(posHadoop, func(i, j int) bool {
		return (posHadoop)[i].Involvement > posHadoop[j].Involvement
	})
	//csvtag.DumpToFile(posHadoop, "C:\\Users\\MasterCai\\Desktop\\5_2\\Hadoop.csv")
	return pps
}

func ProjectPersonByProject(db *gorm.DB, project string) (ProjectPerson, []PersonOperation) {
	pos := PersonOperationByProject(db, project)
	pp := ProjectPerson{}
	pp.Name = pro.GetNameByKey(project)
	creators, assignees, modifiers := getCreatorsAssigneesModifiers(pos)
	calProjectCreator(&pp, creators)
	calProjectAssignee(&pp, assignees)
	calProjectModifier(&pp, modifiers)
	calProjectIncDecRe(&pp, modifiers)
	calProjectPercent(&pp)
	return pp, pos
}

func getCreatorsAssigneesModifiers(pos []PersonOperation) (creators, assignees, modifiers []PersonOperation) {
	creators = make([]PersonOperation, 0)
	assignees = make([]PersonOperation, 0)
	modifiers = make([]PersonOperation, 0)
	for _, po := range pos {
		if po.CreateCount >= 10 {
			creators = append(creators, po)
		}
		if po.AssigneeCount >= 10 {
			assignees = append(assignees, po)
		}
		if po.ModifyCount > 0 {
			modifiers = append(modifiers, po)
		}
	}
	return
}

func calProjectCreator(pp *ProjectPerson, creators []PersonOperation) {
	f := make([]float64, 0)
	for _, creator := range creators {
		f = append(f, creator.CreatePercent)
	}
	m := mean(f)
	s := std(f)
	for _, p := range f {
		if p > (m + s) {
			pp.CreatePreferC++
		}
	}
	pp.CreateTotal = len(f)
}

func calProjectAssignee(pp *ProjectPerson, assignees []PersonOperation) {
	f := make([]float64, 0)
	for _, assignee := range assignees {
		f = append(f, assignee.AssigneePercent)
	}
	m := mean(f)
	s := std(f)
	for _, p := range f {
		if p > (m + s) {
			pp.AssigneePreferC++
		}
	}
	pp.AssigneeTotal = len(f)
}

func calProjectModifier(pp *ProjectPerson, modifiers []PersonOperation) {
	f := make([]float64, 0)
	for _, modifier := range modifiers {
		f = append(f, modifier.ModifyPercent)
	}
	m := mean(f)
	s := std(f)
	for _, p := range f {
		if p > (m + s) {
			pp.ModifyPreferC++
		}
	}
	pp.ModifyTotal = len(f)
}

func calProjectIncDecRe(pp *ProjectPerson, modifiers []PersonOperation) {
	for _, modifier := range modifiers {
		if modifier.IncModifyCount > modifier.DecModifyCount {
			pp.IncPreferC++
		} else if modifier.IncModifyCount < modifier.DecModifyCount {
			pp.DecPreferC++
		} else {
			pp.RePreferC++
		}
	}
}

func calProjectPercent(pp *ProjectPerson) {
	pp.CreatePreferP = RoundN(100*float64(pp.CreatePreferC)/float64(pp.CreateTotal), 1)
	pp.AssigneePreferP = RoundN(100*float64(pp.AssigneePreferC)/float64(pp.AssigneeTotal), 1)
	pp.ModifyPreferP = RoundN(100*float64(pp.ModifyPreferC)/float64(pp.ModifyTotal), 1)
	t := float64(pp.ModifyTotal)
	pp.IncPreferP = RoundN(100*float64(pp.IncPreferC)/t, 1)
	pp.DecPreferP = RoundN(100*float64(pp.DecPreferC)/t, 1)
	pp.RePreferP = RoundN(100*float64(pp.RePreferC)/t, 1)
}

// PersonOperationByProject 计算该项目每个人的三种情况次数
func PersonOperationByProject(db *gorm.DB, project string) []PersonOperation {
	pos := make([]PersonOperation, 0)
	pcMap := QueryAllIssue(db)
	records := issue.QueryBugsByProject(db, project)
	pcRecords := QueryBugRecordsByKeyPrefix(db, project)
	pcRecordsMap := recordsToKeyRecordsMap(pcRecords)
	calCreate(&pos, pcMap, records)
	calAssignee(&pos, pcRecordsMap, records)
	calModify(&pos, pcRecords)
	calInvolvement(db, &pos, records)
	filteredPos(&pos)
	calPOPercent(&pos, project)
	return pos
}

//计算每个人创建了多少Bug，其中多少Bug的优先级后来被修改过
func calCreate(pos *[]PersonOperation, pcMap map[string]bool, records []issue.Issue) {
	for _, record := range records {
		po := &(*pos)[getIndexByName(pos, record.Creator)]
		po.CreateCount++
		po.AssigneeCount++
		if _, ok := pcMap[record.Key]; ok {
			po.CreateCountC++
		}
	}
}

//计算每个人分配了多少次优先级，其中又有多少次后来被修改了
func calAssignee(pos *[]PersonOperation, pcRecordsMap map[string][]PriorityChanged, records []issue.Issue) {
	var po *PersonOperation
	for key, pcRecords := range pcRecordsMap {
		if len(pcRecords) == 1 {
			for _, record := range records {
				if key == record.Key {
					po = &(*pos)[getIndexByName(pos, record.Creator)]
					po.AssigneeCountC++
				}
			}
		} else {
			for i := 0; i < len(pcRecords)-1; i++ {
				po = &(*pos)[getIndexByName(pos, pcRecords[i].Author)]
				po.AssigneeCount++
				po.AssigneeCountC++
			}
		}
		po = &(*pos)[getIndexByName(pos, pcRecords[len(pcRecords)-1].Author)]
		po.AssigneeCount++
	}
}

//计算每个人修改优先级的次数
func calModify(pos *[]PersonOperation, pcRecords []PriorityChanged) {
	personRangeMap := make(map[string]float64)
	for _, record := range pcRecords {
		po := &(*pos)[getIndexByName(pos, record.Author)]
		po.ModifyCount++
		to, _ := strconv.Atoi(toId(record.To))
		from, _ := strconv.Atoi(toId(record.From))
		if to < from {
			po.IncModifyCount++
		} else {
			po.DecModifyCount++
		}
		abs := math.Abs(float64(to) - float64(from))
		personRangeMap[record.Author] += abs
	}
}

//计算每个人参与了多少个Bug,并计算修改率
func calInvolvement(db *gorm.DB, pos *[]PersonOperation, records []issue.Issue) {
	for _, record := range records {
		personMap := make(map[string]bool)
		personMap[record.Creator] = true
		personMap[record.Assignee] = true
		commentersMap := comment.QueryCommentersByRowIds(db, record.Comments)
		for k := range commentersMap {
			personMap[k] = true
		}
		modifiersMap := his.QueryFieldModifiersByRowIds(db, record.HistoryIds)
		for k := range modifiersMap {
			personMap[k] = true
		}
		for person := range personMap {
			(*pos)[getIndexByName(pos, person)].Involvement++
		}
	}
}

//过滤
func filteredPos(pos *[]PersonOperation) {
	for i := 0; i < len(*pos); i++ { //过滤掉3个都为0的人
		if (*pos)[i].AssigneeCount == 0 {
			*pos = append((*pos)[:i], (*pos)[i+1:]...)
			i--
		}
	}
	for i := 0; i < len(*pos); i++ { //删掉“”，这是没有人名的
		if (*pos)[i].Name == "" {
			*pos = append((*pos)[:i], (*pos)[i+1:]...)
			break
		}
	}
	sort.Slice(*pos, func(i, j int) bool {
		return (*pos)[i].Involvement > (*pos)[j].Involvement
	})
	tInv := 0
	for _, po := range *pos {
		tInv += po.Involvement
	}
	Inv80 := 0
	for i := 0; i < len(*pos); i++ {
		Inv80 += (*pos)[i].Involvement
		if float64(Inv80) > float64(tInv)*0.8 {
			*pos = (*pos)[:i+1]
			break
		}
	}
	for _, po := range *pos {
		totalCre += po.CreateCount
		totalAss += po.AssigneeCount
		totalInv += po.Involvement
	}
	lenTotal += len(*pos)
}

//计算百分比
func calPOPercent(pos *[]PersonOperation, project string) {
	for i := 0; i < len(*pos); i++ {
		(*pos)[i].CreatePercent = float64((*pos)[i].CreateCountC) / float64((*pos)[i].CreateCount)
		(*pos)[i].AssigneePercent = float64((*pos)[i].AssigneeCountC) / float64((*pos)[i].AssigneeCount)
		(*pos)[i].ModifyPercent = float64((*pos)[i].ModifyCount) / float64((*pos)[i].Involvement)
		if (*pos)[i].CreatePercent >= 0.166 && (*pos)[i].CreateCount >= 38 {
			fmt.Println(project, "Cr", (*pos)[i].Name, (*pos)[i].CreateCount, (*pos)[i].CreateCountC, (*pos)[i].CreatePercent)
		} else if (*pos)[i].AssigneePercent >= 0.166 && (*pos)[i].AssigneeCount >= 44 {
			fmt.Println(project, "As", (*pos)[i].Name, (*pos)[i].AssigneeCount, (*pos)[i].AssigneeCountC, (*pos)[i].AssigneePercent)
		} else if (*pos)[i].ModifyPercent >= 0.166 && (*pos)[i].Involvement >= 145 {
			fmt.Println(project, "Md", (*pos)[i].Name, (*pos)[i].Involvement, (*pos)[i].ModifyCount, (*pos)[i].ModifyPercent)
		}
	}
	totalCre = 0
	totalAss = 0
	totalInv = 0
	lenTotal = 0
}

func getIndexByName(pos *[]PersonOperation, name string) int {
	for i, po := range *pos {
		if po.Name == name {
			return i
		}
	}
	po := &PersonOperation{}
	po.Name = name
	*pos = append(*pos, *po)
	return len(*pos) - 1
}
