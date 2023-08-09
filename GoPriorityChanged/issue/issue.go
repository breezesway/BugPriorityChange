package issue

import (
	"gorm.io/gorm"
	"sync"
)

type Issue struct {
	Key            string `gorm:"column:key"`
	Created        string `gorm:"column:created"`
	ResolutionDate string `gorm:"column:resolutiondate"`
	Resolution     string `gorm:"column:resolution"`
	Issuetype      string `gorm:"column:issuetype"`
	Status         string `gorm:"column:status"`
	Project        string `gorm:"column:project"`
	Priority       string `gorm:"column:priority"`
	HistoryIds     string `gorm:"column:histories"`
	Comments       string `gorm:"column:comments"`
	Creator        string `gorm:"column:creatordisplayname"`
	Assignee       string `gorm:"column:assigneedisplayname"`
}

// InitPriorityModifiedMap 记录初始优先级被修正了的Bug
var InitPriorityModifiedMap = make(map[string]string)

// ProjectBugCacheMap 缓存项目的Bug
var ProjectBugCacheMap = make(map[string][]Issue)
var mutex sync.Mutex

// QueryProjectBugCount 查询项目的Bug数量
func QueryProjectBugCount(db *gorm.DB, project string) int {
	return len(QueryBugsByProject(db, project))
}

// QueryBugsByProject 查询该项目的Bug
func QueryBugsByProject(db *gorm.DB, project string) []Issue {
	if _, ok := ProjectBugCacheMap[project]; ok {
		return ProjectBugCacheMap[project]
	}
	var isssues []Issue
	db.Table("issue").Select("issue.key,created,resolutiondate,resolution,issuetype,status,priority,project,assigneedisplayname,creatordisplayname,comments,histories").Where("project = ? AND issuetype = ?", project, "Bug").Find(&isssues)
	for i := 0; i < len(isssues); i++ {
		if p, ok := InitPriorityModifiedMap[isssues[i].Key]; ok {
			isssues[i].Priority = p
		}
		p := isssues[i].Priority
		if p != "Blocker" && p != "Critical" && p != "Major" && p != "Minor" && p != "Trivial" {
			isssues = append(isssues[:i], isssues[i+1:]...)
			i--
		}
	}
	mutex.Lock()
	if _, ok := ProjectBugCacheMap[project]; !ok {
		ProjectBugCacheMap[project] = isssues
	}
	mutex.Unlock()
	return isssues
}

// QueryIssueByKey 根据Key查询该Issue
func QueryIssueByKey(db *gorm.DB, key string) Issue {
	var issue Issue
	db.Table("issue").Select("issue.key", "issuetype", "created", "assigneedisplayname", "creatordisplayname", "comments", "histories").Where("issue.key = ?", key).Find(&issue)
	return issue
}
