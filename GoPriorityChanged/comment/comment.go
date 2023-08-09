package comment

import (
	"gorm.io/gorm"
	"strings"
)

type Comment struct {
	Body   string `gorm:"column:body"`
	Author string `gorm:"column:authordisplayname"`
}

// QueryCommentById 根据ID查询该条评论
func QueryCommentById(db *gorm.DB, id string) Comment {
	var comment Comment
	db.Table("comment").Select("body", "authordisplayname").Where("id = ?", id).Find(&comment)
	return comment
}

// QueryCommentersByRowIds 根据原始Id数组查询并返回所有评论者集合
func QueryCommentersByRowIds(db *gorm.DB, rowCommentIds string) map[string]bool {
	s := rowCommentIds[1 : len(rowCommentIds)-1]
	commentIds := strings.Split(s, ", ")
	authorMap := make(map[string]bool)
	for _, id := range commentIds {
		c := QueryCommentById(db, id)
		authorMap[c.Author] = true
	}
	return authorMap
}
