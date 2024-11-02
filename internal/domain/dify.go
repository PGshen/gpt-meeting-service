package domain

// 应用类型
type DifyAppType string

const (
	ChatBot        DifyAppType = "chatBot"
	Agent          DifyAppType = "agent"
	TextCompletion DifyAppType = "textCompletion"
	Workflow       DifyAppType = "workflow"
)

// 分类
type DifyCategory string

const (
	Learning    DifyCategory = "learning"
	Efficiency  DifyCategory = "efficiency"
	Programming DifyCategory = "programming"
	Business    DifyCategory = "business"
	Writing     DifyCategory = "writing"
	Image       DifyCategory = "image"
	Audio       DifyCategory = "audio"
	Video       DifyCategory = "video"
	Characters  DifyCategory = "characters"
	Lifestyle   DifyCategory = "lifestyle"
	Games       DifyCategory = "games"
	Other       DifyCategory = "other"
)

// 排序
type DifySort string

const (
	Like       DifySort = "like"
	Download   DifySort = "download"
	CreateTime DifySort = "createTime"
)

// dify 模板
type Dify struct {
	Id          string         `bson:"_id,omitempty"`
	Name        string         `bson:"name"`
	Description string         `bson:"description"`
	Author      string         `bson:"author"`
	AppType     DifyAppType    `bson:"appType"`
	Category    []DifyCategory `bson:"category"`
	Yml         string         `bson:"yml"`
	Images      []string       `bson:"images"`
	DownloadCnt int64          `bson:"downloadCnt"`
	DownloadIps []string       `bson:"downloadIps"`
	LikeCnt     int64          `bson:"likeCnt"`
	LikeIps     []string       `bson:"likeIps"`
	DislikeCnt  int64          `bson:"dislikeCnt"`
	DislikeIps  []string       `bson:"dislikeIps"`
	Quality     int64          `bson:"quality"`
	CreateTime  int64          `bson:"createTime"`
	UpdateTime  int64          `bson:"updateTime"`
	DeleteTime  int64          `bson:"deleteTime"`
}

// 查询请求参数
type DifySearchReq struct {
	PageNum  int64        `json:"pageNum"`
	Name     string       `json:"name"`
	AppType  DifyAppType  `json:"appType"`
	Category DifyCategory `json:"category"`
	Sort     DifySort     `json:"sort"`
}

type DifyData struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	AppType     DifyAppType    `json:"appType"`
	Category    []DifyCategory `json:"category"`
	Yml         string         `json:"yml"`
	Images      []string       `json:"images"`
	DownloadCnt int64          `json:"downloadCnt"`
	LikeCnt     int64          `json:"likeCnt"`
}

type DifyResp struct {
	Cnt  int64       `json:"cnt"`
	List *[]DifyData `json:"list"`
}

type DifyEmpty struct{}
