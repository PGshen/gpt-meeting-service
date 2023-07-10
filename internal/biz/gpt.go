/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-21 12:03:38
 * @LastEditTime: 2023-07-03 21:06:49
 */
package biz

import (
	"context"
	"errors"
	"fmt"
	"gpt-meeting-service/internal/conf"
	"io"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/sashabaranov/go-openai"
)

type Gpt struct {
	data *conf.Data
	log  *log.Helper
}

func NewGpt(data *conf.Data, logger log.Logger) *Gpt {
	return &Gpt{
		data: data,
		log:  log.NewHelper(logger),
	}
}

/**
 * @description: 请求GPT
 * @param {http.ResponseWriter} resp
 * @param {string} apiKey
 * @param {string} model
 * @param {[]openai.ChatCompletionMessage} message
 * @param {float32} temperature
 * @return {*}
 */
func (g *Gpt) ChatCompletion(ctx http.Context, message []openai.ChatCompletionMessage) (gptReply string, err error) {
	headers := ctx.Request().Header
	resp := ctx.Response()
	apiKey := g.data.OpenAi.ApiKey
	if len(headers["Apikey"]) > 0 && headers["Apikey"][0] != "" {
		apiKey = headers["Apikey"][0]
	}
	model := g.data.OpenAi.Model
	if len(headers["Model"]) > 0 && headers["Model"][0] != "" {
		model = headers["Model"][0]
	}
	temperature := g.data.OpenAi.Temperature
	if len(headers["Temperature"]) > 0 && headers["Temperature"][0] != "0" {
		temp, _ := strconv.ParseFloat(headers["Temperature"][0], 32)
		temperature = float32(temp)
	}
	presencePenalty := g.data.OpenAi.PresencePenalty
	if len(headers["Presencepenalty"]) > 0 && headers["Presencepenalty"][0] != "0" {
		temp, _ := strconv.ParseFloat(headers["Presencepenalty"][0], 32)
		presencePenalty = float32(temp)
	}

	c := openai.NewClient(apiKey)
	openaiReq := openai.ChatCompletionRequest{
		Model:           model,
		Messages:        message,
		Temperature:     temperature,
		PresencePenalty: presencePenalty,
		Stream:          true,
	}
	stream, err := c.CreateChatCompletionStream(context.Background(), openaiReq)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return "", err
	}
	defer stream.Close()
	defer func() { g.printLog(message, gptReply) }()

	fmt.Printf("Stream response: ")
	// gptReply := ""
	index := 1
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			resp.Write([]byte("id: " + strconv.Itoa(index) + "\nevent: close\ndata: close\n\n"))
			return gptReply, err
		}

		if err != nil {
			return "", err
		}

		content := response.Choices[0].Delta.Content
		fmt.Print(content)
		gptReply += content
		content = strings.ReplaceAll(content, "\n", "\ndata: ") // 换行符替换成data:
		sendContent := fmt.Sprintf("event: %s\ndata: %s\nid: %d\n\n", "chat", content, index)
		resp.Write([]byte(sendContent))
		if f, ok := resp.(http.Flusher); ok {
			f.Flush()
		}
		index++
	}
}

func (g *Gpt) printLog(messages []openai.ChatCompletionMessage, gptReply string) {
	var gptRequest string
	for _, message := range messages {
		gptRequest += message.Content + "\n"
	}
	fmt.Printf("[GPT]\nRequestMsg -> %s\nReplyMsg -> %s\n[GPT]\n", gptRequest, gptReply)
	g.log.Debugf("[GPT]\nRequestMsg -> %s\nReplyMsg -> %s\n[GPT]\n", gptRequest, gptReply)
}
