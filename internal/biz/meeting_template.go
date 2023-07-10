/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-03 14:14:57
 * @LastEditTime: 2023-07-09 21:09:58
 */
package biz

import (
	"context"
	"encoding/json"
	"gpt-meeting-service/internal/domain"
	"gpt-meeting-service/internal/lib/datastruct"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type MeetingTemplateRepo interface {
	CreateMeetingTemplate(context.Context, *domain.MeetingTemplate) (bool, error)
	UpdateMeetingTemplate(context.Context, *domain.MeetingTemplate) (bool, error)
	DeleteMeetingTemplate(context.Context, string) (bool, error)
	List(context.Context, *domain.ListMeetingTemplateCondition) (*domain.MeetingTemplateList, int64, error)
	FindOne(context.Context, string) (*domain.MeetingTemplate, error)
}

type MeetingTemplateUsecase struct {
	repo     MeetingTemplateRepo
	roleRope RoleTemplateRepo
	log      *log.Helper
}

func NewMeetingTemplateUsecase(repo MeetingTemplateRepo, roleRope RoleTemplateRepo, logger log.Logger) *MeetingTemplateUsecase {
	return &MeetingTemplateUsecase{
		repo:     repo,
		roleRope: roleRope,
		log:      log.NewHelper(logger),
	}
}

func (mu *MeetingTemplateUsecase) Create(ctx context.Context, meeting *domain.MeetingTemplate) (bool, error) {
	meeting.CreatedTime = time.Now().Unix()
	// 解析图
	graph := datastruct.NewGraph[domain.MeetingNode]()
	var templateFlow map[string]([]map[string]interface{})
	if err := json.Unmarshal([]byte(meeting.TemplateFlow), &templateFlow); err != nil {
		return false, err
	}
	var templateData map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(meeting.TemplateData), &templateData); err != nil {
		return false, err
	}
	cells := templateFlow["cells"]
	edgeList := make([]map[string]string, 0)
	nodeMap := make(map[string]domain.MeetingNode)
	for _, cell := range cells {
		shape := cell["shape"]
		if shape == "data-edge" {
			// 边
			source := cell["source"].(map[string]interface{})
			target := cell["target"].(map[string]interface{})
			// fmt.Printf("%T", source["cell"])
			edge := map[string]string{
				"source": source["cell"].(string),
				"target": target["cell"].(string),
			}
			edgeList = append(edgeList, edge)
		} else if shape == "data-node" {
			// 节点
			id := cell["id"].(string)
			data := templateData[id]
			node := domain.MeetingNode{
				Id:       id,
				NodeType: data["nodeType"].(string),
				NodeName: data["nodeName"].(string),
			}
			if data["characters"] != nil {
				memberId := data["characters"].(string)
				role, err := mu.roleRope.FindOne(ctx, memberId)
				if err != nil {
					return false, err
				}
				node.Characters = &domain.Member{
					MemberId:    role.Id.String(),
					MemberName:  role.Summary,
					Description: role.Description,
				}
			}
			if data["associationCharacters"] != nil {
				memberId := data["associationCharacters"].(string)
				role, err := mu.roleRope.FindOne(ctx, memberId)
				if err != nil {
					return false, err
				}
				node.AssociationCharacters = &domain.Member{
					MemberId:    role.Id.String(),
					MemberName:  role.Summary,
					Description: role.Description,
				}
			}
			if data["quizCharacters"] != nil {
				memberId := data["quizCharacters"].(string)
				role, err := mu.roleRope.FindOne(ctx, memberId)
				if err != nil {
					return false, err
				}
				node.QuizCharacters = &domain.Member{
					MemberId:    role.Id.String(),
					MemberName:  role.Summary,
					Description: role.Description,
				}
			}
			if data["memberIds"] != nil {
				memberList := []domain.Member{}
				temp := data["memberIds"].([]interface{})
				for _, memberId := range temp {
					role, err := mu.roleRope.FindOne(ctx, memberId.(string))
					if err != nil {
						return false, err
					}
					memberList = append(memberList, domain.Member{
						MemberId:    role.Id.String(),
						MemberName:  role.Summary,
						Description: role.Description,
					})
				}
				node.MemberList = memberList
			}
			if data["optimizationPrompt"] != nil {
				node.OptimizationPrompt = data["optimizationPrompt"].(string)
			}
			if data["prologuePrompt"] != nil {
				node.ProloguePrompt = data["prologuePrompt"].(string)
			}
			if data["summaryPrompt"] != nil {
				node.SummaryPrompt = data["summaryPrompt"].(string)
			}
			if data["processingPrompt"] != nil {
				node.ProcessingPrompt = data["processingPrompt"].(string)
			}
			if data["quizPrompt"] != nil {
				node.QuizPrompt = data["quizPrompt"].(string)
			}
			if data["quizRound"] != nil {
				node.QuizRound = int32(data["quizRound"].(float64))
			}
			if data["quizNum"] != nil {
				node.QuizNum = int32(data["quizNum"].(float64))
			}
			if data["replyRound"] != nil {
				node.ReplyRound = int32(data["replyRound"].(float64))
			}
			graph.AddNode(node)
			nodeMap[id] = node
		}
	}
	for _, edge := range edgeList {
		graph.AddEdge(nodeMap[edge["source"]], nodeMap[edge["target"]])
	}
	meeting.TemplateGraph = *graph
	return mu.repo.CreateMeetingTemplate(ctx, meeting)
}

func (mu *MeetingTemplateUsecase) Update(ctx context.Context, meeting *domain.MeetingTemplate) (bool, error) {
	return mu.repo.UpdateMeetingTemplate(ctx, meeting)
}

func (mu *MeetingTemplateUsecase) Delete(ctx context.Context, id string) (bool, error) {
	return mu.repo.DeleteMeetingTemplate(ctx, id)
}

func (mu *MeetingTemplateUsecase) Find(ctx context.Context, condition *domain.ListMeetingTemplateCondition) (*domain.MeetingTemplateList, int64, error) {
	return mu.repo.List(ctx, condition)
}

func (mu *MeetingTemplateUsecase) FindById(ctx context.Context, id string) (*domain.MeetingTemplate, error) {
	return mu.repo.FindOne(ctx, id)
}
