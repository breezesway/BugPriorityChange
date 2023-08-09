package pro

import (
	"encoding/csv"
	"log"
	"os"
	"sort"
	"strings"
)

// ReadProjects 从csv文件中读取项目名
func ReadProjects(path string) []string {
	var projects []string
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("can not open the file, err is %+v", err)
	}
	defer file.Close()
	r := csv.NewReader(file)
	contenet, err := r.ReadAll()
	if err != nil {
		log.Fatalf("can not readall, err is %+v", err)
	}
	for _, row := range contenet {
		projects = append(projects, row[0])
	}
	projects = projects[1:]
	sort.Slice(projects, func(i, j int) bool {
		c := strings.Compare(GetNameByKey(projects[i]), GetNameByKey(projects[j]))
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	return projects
}
