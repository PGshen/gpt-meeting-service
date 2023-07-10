/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-03 14:26:02
 * @LastEditTime: 2023-07-09 21:08:18
 */
package domain

import (
	pb "gpt-meeting-service/api/template/v1"
	"gpt-meeting-service/internal/lib/datastruct"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
 * 节点类型枚举
 */
const (
	NT_Input      string = "Input"
	NT_Output     string = "Output"
	NT_Thinking   string = "Thinking"
	NT_Discussion string = "Discussion"
	NT_Common     string = "Common"
)

type MeetingTemplate struct {
	Id            primitive.ObjectID            `bson:"_id,omitempty"`
	Name          string                        `bson:"name"`
	Avatar        string                        `bson:"avatar"`
	Description   string                        `bson:"description"`
	Example       string                        `bson:"example"`
	TemplateFlow  string                        `bson:"templateFlow"`
	TemplateData  string                        `bson:"templateData"`
	TemplateGraph datastruct.Graph[MeetingNode] `bson:"templateGraph"`
	StarCount     int64                         `bson:"starCount"`
	CreatedBy     string                        `bson:"createdBy"`
	CreatedTime   int64                         `bson:"createdTime"`
}

type MeetingTemplateList []MeetingTemplate

type MeetingNode struct {
	Id                    string                 `bson:"_id" json:"id"`
	NodeType              string                 `bson:"nodeType" json:"nodeType"`                                               // 节点类型
	NodeName              string                 `bson:"nodeName" json:"nodeName"`                                               // 节点名称
	Characters            *Member                `bson:"characters,omitempty" json:"characters,omitempty"`                       // AI人设
	AssociationCharacters *Member                `bson:"associationCharacters,omitempty" json:"associationCharacters,omitempty"` // 联想AI人设-thinking专有
	QuizCharacters        *Member                `bson:"quizCharacters,omitempty" json:"quizCharacters,omitempty"`               // 提问AI人设-thinking专有
	OptimizationPrompt    string                 `bson:"optimizationPrompt,omitempty" json:"optimizationPrompt,omitempty"`       // 优化prompt
	ProloguePrompt        string                 `bson:"prologuePrompt,omitempty" json:"prologuePrompt,omitempty"`               // 开场prompt
	SummaryPrompt         string                 `bson:"summaryPrompt,omitempty" json:"summaryPrompt,omitempty"`                 // 总结prompt
	ProcessingPrompt      string                 `bson:"processingPrompt,omitempty" json:"processingPrompt,omitempty"`           // 处理prompt
	QuizPrompt            string                 `bson:"quizPrompt,omitempty" json:"quizPrompt,omitempty"`                       // 提问prompt
	QuizRound             int32                  `bson:"quizRound,omitempty" json:"quizRound,omitempty"`                         // 提问轮次
	QuizNum               int32                  `bson:"quizNum,omitempty" json:"quizNum,omitempty"`                             // 提问次数
	ReplyRound            int32                  `bson:"replyRound,omitempty" json:"replyRound,omitempty"`                       // 回复次数
	MemberList            []Member               `bson:"memberList,omitempty" json:"memberList,omitempty"`                       // 参与本环节的成员（角色ID）列表
	OtherInfo             map[string]interface{} `bson:"otherInfo,omitempty" json:"otherInfo"`                                   // 其他信息

}

func (m MeetingNode) GetId() string {
	return m.Id
}

// 成员
type Member struct {
	MemberId    string `bson:"memeberId" json:"memberId"`
	MemberName  string `bson:"memberName" json:"memberName"`
	Description string `bson:"description" json:"description"`
}

type ListMeetingTemplateCondition struct {
	Id       string
	Name     string
	PageNum  int64
	PageSize int64
}

func (domain *MeetingTemplate) MeetingTemplateToPb() *pb.MeetingInfo {
	return &pb.MeetingInfo{
		Id:           domain.Id.Hex(),
		Name:         domain.Name,
		Avatar:       domain.Avatar,
		Description:  domain.Description,
		Example:      domain.Example,
		TemplateFlow: domain.TemplateFlow,
		TemplateData: domain.TemplateData,
		StarCount:    domain.StarCount,
		CreatedBy:    domain.CreatedBy,
		CreatedTime:  domain.CreatedTime,
	}
}
