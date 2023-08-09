package pri

import (
	"GoPriorityChanged/issue"
	"gorm.io/gorm"
	"strings"
)

type PriorityChanged struct {
	Author  string `gorm:"column:authordisplayname"`
	Created string `gorm:"column:created"`
	Key     string `gorm:"column:issuekey"`
	From    string `gorm:"column:fromstring"`
	To      string `gorm:"column:tostring"`
}

var filteredRecords []PriorityChanged
var RecordsMap map[string][]PriorityChanged

// QueryAllIssue 获取所有发生过优先级变更的Issue
func QueryAllIssue(db *gorm.DB) map[string]bool {
	var records []PriorityChanged
	if len(filteredRecords) != 0 {
		records = filteredRecords
	} else {
		db.Table("prioritychanged").Select("issuekey").Find(&records)
	}
	m := make(map[string]bool)
	for _, record := range records {
		key := record.Key
		if _, ok := m[key]; !ok {
			m[key] = true
		}
	}
	return m
}

// QueryBugRecordsByKeyPrefix 查询该前缀发生优先级改变的Bug记录
func QueryBugRecordsByKeyPrefix(db *gorm.DB, project string) []PriorityChanged {
	var records []PriorityChanged
	if len(filteredRecords) != 0 {
		records = make([]PriorityChanged, 0)
		for _, record := range filteredRecords {
			if strings.HasPrefix(record.Key, project+"-") {
				records = append(records, record)
			}
		}
		return records
	} else {
		db.Table("prioritychanged").
			Where("issuekey LIKE ? AND fromstring IN ('Blocker','Critical','Major','Minor','Trivial') AND tostring IN ('Blocker','Critical','Major','Minor','Trivial')", project+"%").
			Find(&records)
		var fRecords []PriorityChanged
		for _, record := range records {
			if issue.QueryIssueByKey(db, record.Key).Issuetype == "Bug" {
				fRecords = append(fRecords, record)
			}
		}
		return fRecords
	}
}

// SetFilteredBugRecords 获取所有过滤后的Bug类型的优先级变更记录
func SetFilteredBugRecords(db *gorm.DB, interval1, interval2, interval3 float64) {
	var records []PriorityChanged
	db.Table("prioritychanged").
		//Where("fromstring IN ('Blocker','Critical','Major','Minor','Trivial') AND tostring IN ('Blocker','Critical','Major','Minor','Trivial')").
		Find(&records)
	tempRecords := make([]PriorityChanged, 0)
	RecordsMap = recordsToKeyRecordsMap(records)
	for k, v := range RecordsMap { //过滤掉不是5种优先级的issue
		for _, pc := range v {
			f := pc.From
			t := pc.To
			if (f != "Blocker" && f != "Critical" && f != "Major" && f != "Minor" && f != "Trivial") ||
				(t != "Blocker" && t != "Critical" && t != "Major" && t != "Minor" && t != "Trivial") {
				delete(RecordsMap, k)
				break
			}
		}
	}
	for k, v := range RecordsMap { //过滤和处理3个interval
		iss := issue.QueryIssueByKey(db, k)
		if iss.Issuetype == "Bug" {
			if len(v) == 1 && timeSub(v[0].Created, iss.Created).Minutes() < interval1 && v[0].Author == iss.Creator {
				issue.InitPriorityModifiedMap[k] = v[0].To
				v = v[1:]
			} else {
				for i := 0; i+1 < len(v); i++ {
					if i == 0 && timeSub(v[0].Created, iss.Created).Minutes() < interval1 && v[0].Author == iss.Creator {
						issue.InitPriorityModifiedMap[k] = v[0].To
						v = v[1:]
						i--
						continue
					}
					if v[i].To == v[i+1].From && v[i].From == v[i+1].To && v[i].Author == v[i+1].Author {
						if timeSub(v[i+1].Created, v[i].Created).Minutes() < interval2 {
							v = append(v[:i], v[i+2:]...)
							i--
						}
					} else if v[i].To == v[i+1].From && v[i].From != v[i+1].To && v[i].Author == v[i+1].Author {
						if timeSub(v[i+1].Created, v[i].Created).Minutes() < interval3 {
							pc := PriorityChanged{
								v[i+1].Author,
								v[i+1].Created,
								v[i+1].Key,
								v[i].From,
								v[i+1].To,
							}
							v[i] = pc
							v = append(v[:i+1], v[i+2:]...)
							i--
						}
					}
				}
			}
			tempRecords = append(tempRecords, v...)
		}
	}
	filteredRecords = tempRecords
	RecordsMap = recordsToKeyRecordsMap(tempRecords)
}
