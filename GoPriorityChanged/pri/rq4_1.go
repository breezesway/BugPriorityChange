package pri

import (
	"GoPriorityChanged/issue"
	"GoPriorityChanged/pro"
	"errors"
	"fmt"
	csvtag "github.com/artonge/go-csv-tag/v2"
	"gorm.io/gorm"
	"io/ioutil"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ProjectIndex struct {
	pc   ProjectComplexity
	BugC []BugComplexity
	BugN []BugComplexity
}

type ProjectComplexity struct {
	Name        string  `csv:"Name"`
	Key         string  `csv:"Key"`
	LOCAveC     float64 `csv:"LOCAveC"`
	LOCAveN     float64 `csv:"LOCAveN"`
	NOFAveC     float64 `csv:"NOFAveC"`
	NOFAveN     float64 `csv:"NOFAveN"`
	NOPAveC     float64 `csv:"NOPAveC"`
	NOPAveN     float64 `csv:"NOPAveN"`
	EntropyAveC float64 `csv:"EntropyAveC"`
	EntropyAveN float64 `csv:"EntropyAveN"`
}

type BugComplexity struct {
	Key     string  `csv:"Key"`
	LOC     float64 `csv:"LOC"`
	NOF     float64 `csv:"NOF"`
	NOP     float64 `csv:"NOP"`
	Entropy float64 `csv:"Entropy"`
}

func RQ4_1(db *gorm.DB, path string, out string) {
	start := time.Now()
	complexities := Complexities(db, path)
	pcs := make([]ProjectComplexity, 0)
	for _, c := range complexities {
		sort.Slice(c.BugC, func(i, j int) bool {
			return c.BugC[i].LOC > c.BugC[j].LOC
		})
		sort.Slice(c.BugN, func(i, j int) bool {
			return c.BugN[i].LOC > c.BugN[j].LOC
		})
		csvtag.DumpToFile(c.BugC, "C:\\Users\\MasterCai\\Desktop\\4_1\\"+c.pc.Name+"C.csv")
		csvtag.DumpToFile(c.BugN, "C:\\Users\\MasterCai\\Desktop\\4_1\\"+c.pc.Name+"N.csv")
		pcs = append(pcs, c.pc)
	}
	sort.Slice(pcs, func(i, j int) bool {
		c := strings.Compare(pcs[i].Name, pcs[j].Name)
		if c < 0 {
			return true
		} else {
			return false
		}
	})
	for _, pc := range pcs {
		if pc.LOCAveN > pc.LOCAveC || pc.NOFAveN > pc.NOFAveC || pc.NOPAveN > pc.NOPAveC || pc.EntropyAveN > pc.EntropyAveC {
			fmt.Println(pc)
		}
	}
	csvtag.DumpToFile(pcs, out)
	fmt.Println("RQ4_1耗时:", time.Since(start).String())
	Wg.Done()
}

// Complexities 获取所有项目的各个指标及其所有Bug的指标
func Complexities(db *gorm.DB, path string) []ProjectIndex {
	pis := make([]ProjectIndex, 0)
	m := GetProComFileMap(path)
	pcMap := QueryAllIssue(db)
	for name, paths := range m { //循环遍历每个name
		keys, err := pro.GetKeyByName(name)
		if err != nil {
			fmt.Println(name, err)
			return nil
		}
		for _, key := range keys { //循环每个key
			index := containKey(pis, key)
			var pi *ProjectIndex
			if index == -1 {
				pi = &ProjectIndex{}
				pi.pc.Name = name
				pi.pc.Key = key
				pis = append(pis, *pi)
			}
			index = containKey(pis, key)
			pi = &pis[index]
			for _, v := range paths { //循环每个commit文件
				lines := ReadCommitFile(v)
				commits := ParseCommit(lines)
				bcMap := bugCommitMap(db, key, commits)
				bugC, bugN := calBugMetric(pcMap, bcMap, commits)
				pi.BugC = append(pi.BugC, bugC...)
				pi.BugN = append(pi.BugN, bugN...)
			}
			pc := CalProjectMetric(pi)
			pc.Name = name
			pc.Key = key
			pi.pc = pc
		}
	}
	hadoop := ProjectIndex{}
	for i := 0; i < len(pis); i++ {
		key := pis[i].pc.Key
		if key == "HADOOP" || key == "HDFS" || key == "MAPREDUCE" || key == "YARN" {
			hadoop.BugC = append(hadoop.BugC, pis[i].BugC...)
			hadoop.BugN = append(hadoop.BugN, pis[i].BugN...)
			pis = append(pis[:i], pis[i+1:]...)
			i--
		}
	}
	hadoop.pc = CalProjectMetric(&hadoop)
	hadoop.pc.Name = "Hadoop"
	hadoop.pc.Key = "HADOOP, MAPREDUCE, YARN, HDFS"
	pis = append(pis, hadoop)
	return pis
}

func Test4_1(db *gorm.DB) ([]BugComplexity, []BugComplexity) {
	lines := ReadCommitFile("E:\\PriorityChange\\ProjectCommit\\lucene_Git_NoMerge_AllBrch_ASC.txt")
	fmt.Println(len(lines))
	commits := ParseCommit(lines)
	fmt.Println(len(commits))
	bcMap := bugCommitMap(db, "LUCENE", commits)
	fmt.Println(len(bcMap))
	m := QueryAllIssue(db)
	bugC, bugN := calBugMetric(m, bcMap, commits)
	fmt.Println(len(bugC), len(bugN))
	var pi *ProjectIndex
	pi = &ProjectIndex{}
	pi.BugC = append(pi.BugC, bugC...)
	pi.BugN = append(pi.BugN, bugN...)
	bugs := append(bugN, bugC...)
	locs := make([]int, 0)
	for _, bug := range bugs {
		locs = append(locs, int(bug.LOC))
	}
	sort.Ints(locs)
	for _, loc := range locs {
		fmt.Println(loc)
	}
	pc := CalProjectMetric(pi)
	pi.pc = pc
	fmt.Println(pi.pc.LOCAveC, pi.pc.LOCAveN, pi.pc.NOFAveC, pi.pc.NOFAveN, pi.pc.NOPAveC, pi.pc.NOPAveN, pi.pc.EntropyAveC, pi.pc.EntropyAveN)
	/*csvtag.DumpToFile(bugC, "C:\\Users\\MasterCai\\Desktop\\BugC_TS.csv")
	csvtag.DumpToFile(bugN, "C:\\Users\\MasterCai\\Desktop\\BugN_TS.csv")*/
	return pi.BugC, pi.BugN
}

//判断是否已经包含某个key，如果是，返回下标，否则返回-1
func containKey(pi []ProjectIndex, key string) int {
	for i, p := range pi {
		if key == p.pc.Key {
			return i
		}
	}
	return -1
}

func containKey2(pi []ProjectMLBugSLBug, key string) int {
	for i, p := range pi {
		if key == p.Key {
			return i
		}
	}
	return -1
}

// 扫描commit切片，返回一个key-对应commit(s)的map，只选择类型为Bug的issue
func bugCommitMap(db *gorm.DB, pre string, commits []Commit) map[string][]int {
	m := make(map[string][]int)
	reg, _ := regexp.Compile(pre + "-\\d+")
	for i, commit := range commits {
		keys := reg.FindAllString(commit.Message, -1)
		if keys != nil {
			for _, key := range keys {
				if issue.QueryIssueByKey(db, key).Issuetype == "Bug" {
					if _, ok := m[key]; !ok {
						m[key] = []int{i}
					} else {
						m[key] = append(m[key], i)
					}
				}
			}
		}
	}
	return m
}

//计算每个Bug的指标，并按是否存在优先级改变分为两列返回
func calBugMetric(pcMap map[string]bool, m map[string][]int, commits []Commit) (BugC, BugN []BugComplexity) {
	if len(m) == 0 {
		return
	}
	for key, indexs := range m {
		c := false
		fMap := make(map[string]bool)
		for _, index := range indexs {
			for _, file := range commits[index].Files {
				fMap[file.Op] = true
			}
		}
		if len(fMap) == 1 {
			for k := range fMap {
				if k == "Added" || k == "Deleted" || k == "Renamed" {
					c = true
				}
			}
		}
		if c {
			continue
		}
		bc := BugComplexity{}
		bc.Key = key
		loc, nof, nop := countLOCNOFNOP(indexs, commits)
		if loc >= 10000 || loc == 0 || nof >= 100 {
			continue
		}
		bc.LOC = loc
		bc.NOF = nof
		bc.NOP = nop
		e := countEntropy(key, indexs, commits)
		bc.Entropy = e
		/*if (e > 0.99 || e < 0.8 || e == 0) && (nof >= 10 || loc > 1000) { // || nof < 10 && bc.Entropy == 1
			continue
		}*/
		if _, ok := pcMap[key]; ok {
			BugC = append(BugC, bc)
		} else {
			BugN = append(BugN, bc)
		}
	}
	return
}

//统计该Bug修改的行数、修改的文件数、修改的包数
func countLOCNOFNOP(indexs []int, commits []Commit) (float64, float64, float64) {
	loc := 0
	fMap := make(map[string]bool)
	pMap := make(map[string]bool)
	//addMap := make(map[string]int)
	for _, i := range indexs {
		for _, fileOp := range commits[i].Files {
			l := fileOp.Add + fileOp.Del
			loc += l
			fMap[fileOp.Name] = true
			fMap[fileOp.OldName] = true
			pMap[fileOp.Pack] = true
			pMap[fileOp.OldPack] = true
			/*if fileOp.Op == "Added" {
				addMap[fileOp.Name] = l
			} else if fileOp.Op == "Deleted" {
				if ol, ok := addMap[fileOp.Name]; ok {
					loc -= ol
					loc -= l
					delete(fMap, fileOp.Name)
					delete(pMap, fileOp.Pack)
				}
			}*/
		}
	}
	return float64(loc), float64(len(fMap)), float64(len(pMap))
}

func containIndexs(indexs []int, tempi int) int {
	for _, index := range indexs {
		if tempi == index {
			return index
		}
	}
	return -1
}

//计算该Bug的熵
func countEntropy(key string, indexs []int, commits []Commit) float64 {
	fMap := make(map[string]int)
	for _, i := range indexs {
		if i == 0 {
			return 0
		}
		for _, fileOp := range commits[i].Files {
			f := fileOp.Name
			for tempi := i - 1; tempi >= 0 && commits[i].Date.Sub(commits[tempi].Date).Hours()/24 <= 60; tempi-- {
				for _, tempFileOp := range commits[tempi].Files {
					if f == tempFileOp.Name && containIndexs(indexs, tempi) == -1 {
						fMap[f]++
					}
				}
			}
		}
	}
	if len(fMap) <= 1 {
		return 0
	}
	ps := make([]float64, 0)
	totalm := 0
	for _, v := range fMap {
		totalm += v
	}
	for _, v := range fMap {
		ps = append(ps, float64(v)/float64(totalm))
	}
	var Hn float64
	for _, p := range ps {
		Hn += p * math.Log2(p)
	}
	Hn = -Hn
	return Hn / math.Log2(float64(len(fMap)))
}

// CalProjectMetric 计算项目整体平均值
func CalProjectMetric(pi *ProjectIndex) ProjectComplexity {
	pc := ProjectComplexity{}
	if len(pi.BugC) != 0 {
		for _, bc := range pi.BugC {
			pc.LOCAveC += bc.LOC
			pc.NOFAveC += bc.NOF
			pc.NOPAveC += bc.NOP
			pc.EntropyAveC += bc.Entropy
		}
		pc.LOCAveC = RoundN(pc.LOCAveC/float64(len(pi.BugC)), 0)
		pc.NOFAveC = RoundN(pc.NOFAveC/float64(len(pi.BugC)), 2)
		pc.NOPAveC = RoundN(pc.NOPAveC/float64(len(pi.BugC)), 2)
		pc.EntropyAveC = RoundN(pc.EntropyAveC/float64(len(pi.BugC)), 2)
	}
	if len(pi.BugN) != 0 {
		for _, bn := range pi.BugN {
			pc.LOCAveN += bn.LOC
			pc.NOFAveN += bn.NOF
			pc.NOPAveN += bn.NOP
			pc.EntropyAveN += bn.Entropy
		}
		pc.LOCAveN = RoundN(pc.LOCAveN/float64(len(pi.BugN)), 0)
		pc.NOFAveN = RoundN(pc.NOFAveN/float64(len(pi.BugN)), 2)
		pc.NOPAveN = RoundN(pc.NOPAveN/float64(len(pi.BugN)), 2)
		pc.EntropyAveN = RoundN(pc.EntropyAveN/float64(len(pi.BugN)), 2)
	}
	return pc
}

// GetProComFileMap 给定路径，获取该文件夹下所有项目的文件路径
func GetProComFileMap(path string) map[string][]string {
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("文件夹读取出错")
		return nil
	}
	m := make(map[string][]string)
	for _, fi := range rd {
		if fi.IsDir() {
			dir, _ := ioutil.ReadDir(path + "\\" + fi.Name() + "\\")
			m[fi.Name()] = make([]string, 0)
			for _, f := range dir {
				m[fi.Name()] = append(m[fi.Name()], path+"\\"+fi.Name()+"\\"+f.Name())
			}
		} else {
			fN := fi.Name()
			proName := fN[:strings.Index(fN, "_")]
			m[proName] = []string{path + "\\" + fN}
		}
	}
	return m
}

// ReadCommitFile 读取整个文件
func ReadCommitFile(path string) []string {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	lines := strings.Split(string(bytes), "\n")
	return lines
}

type Commit struct {
	Revision string
	Author   string
	Date     time.Time
	Message  string
	Files    []FileOp
}

type FileOp struct {
	Op      string
	Name    string
	OldName string
	Pack    string
	OldPack string
	Suf     string
	Add     int
	Del     int
}

// ParseCommit 解析Commit
func ParseCommit(lines []string) []Commit {
	commits := make([]Commit, 0)
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "Revision") {
			c := Commit{}
			c.Revision = lines[i][strings.Index(lines[i], ":")+2 : len(lines[i])-1]
			i++
			c.Author = lines[i][strings.Index(lines[i], ":")+2 : strings.Index(lines[i], "<")-1]
			i++
			t, _ := time.Parse("02/01/2006 15:04:05", lines[i][strings.Index(lines[i], ":")+2:len(lines[i])-1])
			c.Date = t
			i += 2
			c.Message = lines[i][:len(lines[i])-1]
			i += 3
			c.Files = make([]FileOp, 0)
			for {
				i++
				f, err := parseOneFileLine(lines[i])
				if err != nil {
					if err.Error() == "其他行" {
						continue
					} else {
						break
					}
				}
				if _, ok := SufMap[f.Suf]; ok {
					c.Files = append(c.Files, f)
				}
			}
			commits = append(commits, c)
		}
	}
	return commits
}

//解析一行文件的操作
func parseOneFileLine(line string) (FileOp, error) {
	f := FileOp{}
	if strings.HasPrefix(line, "Modified") ||
		strings.HasPrefix(line, "Added") ||
		strings.HasPrefix(line, "Deleted") {
		d1 := strings.Index(line, ":")
		d2 := strings.LastIndex(line, "|")
		d3 := strings.LastIndex(line, " ")
		f.Op = line[:d1]
		f.Name = line[d1+2 : d2]
		index := strings.LastIndex(f.Name, "/")
		if index != -1 {
			f.Pack = f.Name[:index]
		}
		index = strings.LastIndex(f.Name, ".")
		if index != -1 {
			f.Suf = f.Name[index:]
		}
		atoi, _ := strconv.Atoi(line[d2+1 : d3])
		f.Add = atoi
		atoi, _ = strconv.Atoi(line[d3+1 : len(line)-1])
		f.Del = atoi
		return f, nil
	} else if strings.HasPrefix(line, "Renamed") {
		d1 := strings.Index(line, ":")
		d2 := strings.Index(line, " [")
		d3 := strings.LastIndex(line, "]")
		d4 := strings.LastIndex(line, " ")
		f.Op = line[:d1]
		f.Name = line[d1+2 : d2]
		index := strings.LastIndex(f.Name, "/")
		if index != -1 {
			f.Pack = f.Name[:index]
		}
		index = strings.LastIndex(f.Name, ".")
		if index != -1 {
			f.Suf = f.Name[index:]
		}
		f.OldName = line[d2+7 : d3]
		index = strings.LastIndex(f.OldName, "/")
		if index != -1 {
			f.OldPack = f.OldName[:index]
		}
		atoi, _ := strconv.Atoi(line[d3+2 : d4])
		f.Add = atoi
		atoi, _ = strconv.Atoi(line[d4+1 : len(line)-1])
		f.Del = atoi
		return f, nil
	} else if len(line) > 2 {
		return f, errors.New("其他行")
	} else {
		return f, errors.New("空行")
	}
}

var SufMap map[string]string

func initSufMap() {
	SufMap = make(map[string]string)
	SufMap[".java"] = "Java"
	SufMap[".jav"] = "Java"
	SufMap[".j"] = "Java"
	SufMap[".jsp"] = "Java"

	SufMap[".cs"] = "C#"
	SufMap[".go"] = "Go"

	SufMap[".cpp"] = "C/C++"
	SufMap[".hpp"] = "C/C++"
	SufMap[".hxx"] = "C/C++"
	SufMap[".cxx"] = "C/C++"
	SufMap[".c++"] = "C/C++"
	SufMap[".cc"] = "C/C++"
	SufMap[".hh"] = "C/C++"
	SufMap[".h"] = "C/C++"
	SufMap[".c"] = "C/C++"

	SufMap[".clj"] = "Clojure"
	SufMap[".cljs"] = "Clojure"
	SufMap[".cljc"] = "Clojure"

	SufMap[".coffee"] = "CoffeeScript"
	SufMap[".litcoffee"] = "CoffeeScript"

	SufMap[".erl"] = "Erlang"
	SufMap[".hrl"] = "Erlang"

	SufMap[".hs"] = "Haskell"

	SufMap[".js"] = "JavaScript"
	SufMap[".vue"] = "JavaScript"

	SufMap[".ts"] = "TypeScript"

	SufMap[".kt"] = "Kotlin"
	SufMap[".kts"] = "Kotlin"

	SufMap[".pl"] = "Perl"
	SufMap[".prl"] = "Perl"
	SufMap[".pm"] = "Perl"
	SufMap[".perl"] = "Perl"
	SufMap[".p5"] = "Perl"
	SufMap[".ph"] = "Perl"

	SufMap[".m"] = "Objective-C"
	SufMap[".mm"] = "Objective-C++"
	SufMap[".swift"] = "Swift"

	SufMap[".php"] = "PHP"
	SufMap[".php2"] = "PHP"
	SufMap[".php3"] = "PHP"
	SufMap[".php4"] = "PHP"
	SufMap[".php5"] = "PHP"
	SufMap[".php7"] = "PHP"

	SufMap[".py"] = "Python"
	SufMap[".py2"] = "Python"
	SufMap[".ipynb"] = "Python"

	SufMap[".rb"] = "Ruby"
	SufMap[".rbs"] = "Ruby"
	SufMap[".ruby"] = "Ruby"

	SufMap[".scala"] = "Scala"

	SufMap[".lua"] = "Lua"
	SufMap[".R"] = "R"
	SufMap[".groovy"] = "Groovy"
	SufMap[".rs"] = "Rust"

	/*SufMap[".sh"] = true
	SufMap[".bat"] = true
	SufMap[".patch"] = true
	SufMap[".sql"] = true

	SufMap[".xml"] = true
	SufMap[".yaml"] = true
	SufMap[".properties"] = true
	SufMap[".yml"] = true
	SufMap[".html"] = true
	SufMap[".xsl"] = true
	SufMap[".css"] = true
	SufMap[".json"] = true
	SufMap[".dockerfile"] = true
	SufMap[".ftl"] = true
	SufMap[".jsp"] = true
	SufMap[".cmake"] = true

	SufMap[".build"] = true
	SufMap[".sln"] = true
	SufMap[".template"] = true
	SufMap[".am"] = true
	SufMap[".imp"] = true
	SufMap[".rst"] = true
	SufMap[".proto"] = true
	SufMap[".fbs"] = true
	SufMap[".in"] = true
	SufMap[".env"] = true
	SufMap[".csproj"] = true

	SufMap[".txt"] = true
	SufMap[".md"] = true*/
}
