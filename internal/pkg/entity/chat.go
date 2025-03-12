package entity

import (
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatCreateReq struct {
	ProjectId string `json:"project_id" form:"project_id"`
}

type AgentListReq struct {
	UserId string `json:"user_id" form:"user_id"`
}

type AgentPinReq struct {
	ProjectId string `json:"project_id" form:"project_id"`
}

type AgentUnPinReq struct {
	ProjectId string `json:"project_id" form:"project_id"`
}

type ChatCreateRes struct {
	ID        primitive.ObjectID `json:"id"`
	ProjectId string             `json:"project_id"`
}

type ChatUpdateReq struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type ChatUpdateRes struct {
}

type ChatDeleteReq struct {
	ID string `json:"id"`
}

type ChatDeleteRes struct {
}

type ChatDeleteAllReq struct {
}

type ChatDeleteAllRes struct {
}

type ChatListReq struct {
	Page uint64 `json:"page" form:"page"`
	Size uint64 `json:"size" form:"size"`
}

type ChatListRes struct {
	List  []*model.ChatWindow `json:"list"`
	Total int64               `json:"total"`
}

type AgentListRes struct {
	List  []*model.UserPluginDto `json:"list"`
	Total int64                  `json:"total"`
}

type ChatMsg struct {
	Index     uint64            `json:"index"`
	Timestamp model.JSONTime    `json:"timestamp"`
	Role      model.ChatMsgRole `json:"role"`
	Content   any               `json:"content"`
	ProjectId string            `json:"project_id"`
}

type ChatHistoryReq struct {
	ID   string `json:"id" form:"id"`
	Page uint64 `json:"page" form:"page"`
	Size uint64 `json:"size" form:"size"`
}

type ChatHistoryRes struct {
	Title      string         `json:"title"`
	CreateTime model.JSONTime `json:"create_time"`
	ProjectId  string         `json:"project_id"`
	MsgNum     uint64         `json:"msg_num"`
	Msgs       []*ChatMsg     `json:"msgs"`
}

type ChatCompletionsReq struct {
	RelatedMsgIndex int               `json:"related_msg_index"`
	ID              string            `json:"id"`
	ProjectId       string            `json:"project_id"`
	Type            model.ChatMsgRole `json:"type"`
	Msg             string            `json:"msg"`
}

type ChatCompletionsRes struct {
	Msg *ChatMsg `json:"msg"`
}

// generate by https://api.aidocs.chat/docs#tag/doc-search/operation/searchDocuments
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type HistoryMessage struct {
	Role    string    `json:"role"`
	Content []Message `json:"content"`
}

type History struct {
	Messages []HistoryMessage `json:"messages"`
}

type Request struct {
	Messages        []Message `json:"messages"`
	RelatedProjects []string  `json:"relatedProjects"`
	History         []History `json:"history"`
}

type Document struct {
	Content string  `json:"content"`
	Score   float64 `json:"score"`
	ID      string  `json:"id"`
}

type ProjectConfig struct {
	RespondWhenNoDocument   bool     `json:"respondWhenNoDocument"`
	ID                      string   `json:"id"`
	OutputFormat            string   `json:"outputFormat"`
	NumOfDocumentsReturned  float64  `json:"numOfDocumentsReturned"`
	ChunkSize               float64  `json:"chunkSize"`
	MaximumDocumentDistance float64  `json:"maximumDocumentDistance"`
	NoDocumentResponse      string   `json:"noDocumentResponse"`
	ResponseLanguage        string   `json:"responseLanguage"`
	ProjectID               string   `json:"projectId"`
	Categories              []string `json:"categories"`
	OutputSchema            struct{} `json:"outputSchema"`
}

type ToolCall struct {
	Type       string                 `json:"type"`
	ToolCallID string                 `json:"toolCallId"`
	ToolName   string                 `json:"toolName"`
	Args       map[string]interface{} `json:"args"`
}

type ToolResultData struct {
	Resolved bool           `json:"resolved"`
	Result   map[string]any `json:"result"`
}

type ToolResult struct {
	Type       string                 `json:"type"`
	ToolCallID string                 `json:"toolCallId"`
	ToolName   string                 `json:"toolName"`
	Args       map[string]interface{} `json:"args"`
	Result     ToolResultData         `json:"result"`
}

type Usage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

type Object struct {
	View        string   `json:"view"`
	Content     string   `json:"content"`
	Intention   string   `json:"intention"`
	Intent_keys []string `json:"intent_keys"`
}

type Response struct {
	Text          string        `json:"text"`
	Object        Object        `json:"object"`
	Documents     []Document    `json:"documents"`
	ProjectConfig ProjectConfig `json:"projectConfig"`
	AccessCount   float64       `json:"accessCount"`
	ToolCalls     []ToolCall    `json:"toolCalls"`
	ToolResults   []ToolResult  `json:"toolResults"`
	Usage         Usage         `json:"usage"`
}
