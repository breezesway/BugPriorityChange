package main

import (
	"GoPriorityChanged/pri"
	"GoPriorityChanged/pro"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func main() {
	ConnectDB()

	projectpath := "C:\\Users\\MasterCai\\Desktop\\projects.csv"
	commentfilepath := "E:\\PriorityChange\\ProjectCommit"
	outdir := "C:\\Users\\MasterCai\\Desktop"

	projects := pro.ReadProjects(projectpath)

	RQ(projects, commentfilepath, outdir, true)
}

func RQ(projects []string, cfpath string, outDir string, filtered bool) {
	start := time.Now()
	if filtered {
		pri.SetFilteredBugRecords(DB, 5, 5, 5)
	}
	fmt.Println("过滤records耗时：", time.Since(start).String())
	pri.Wg.Add(1)
	/*go pri.RQ1(DB, projects, outDir+"\\RQ1.csv")
	go pri.RQ2_1(DB, projects, outDir+"\\RQ2_1.csv")
	go pri.RQ2_2(DB, projects, outDir+"\\RQ2_2.csv")
	go pri.RQ3_1(DB, projects, outDir+"\\RQ3_1.csv")
	go pri.RQ3_2(DB, projects, outDir)
	go pri.RQ3_3_1(DB, projects, outDir+"\\RQ3_3_1.csv")
	go pri.RQ3_3_2(DB, projects, outDir+"\\RQ3_3_2.csv")
	go pri.RQ3_4_1(DB, projects, outDir+"\\RQ3_4_1.csv")
	go pri.RQ3_4_2(DB, projects, outDir+"\\RQ3_4_2.csv")
	go pri.RQ4_1(DB, cfpath, outDir+"\\RQ4_1.csv")
	go pri.RQ4_2(DB, projects, outDir+"\\RQ4_2.csv")
	go pri.RQ5_1(DB, projects, outDir)*/
	go pri.RQ5_2(DB, projects, outDir+"\\RQ5_2.csv")
	pri.Wg.Wait()
	fmt.Println("共计耗时：", time.Since(start).String())
}

var DB *gorm.DB

func ConnectDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/jira?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败")
	}
}
