package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"

	// "github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
)

// simpleQA 一次轮简单回答（至少有3种调用方式）
func simpleQA() {
	//选择一个大语言模型
	llm, err := ollama.New(ollama.WithModel("llama2")) //本地安装llama2: https://ollama.com/。至少需要3.8G的磁盘空间。在终端跟llama聊天：ollama run llama2
	// llm, err := openai.New() //需要在配置环境变量OPENAI_API_KEY
	if err != nil {
		log.Fatalln(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, "中国的首都是哪里?",
		llms.WithTemperature(0), //温度越高，回答越发散
		// llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 	fmt.Print(string(chunk))
		// 	return nil
		// }), //流式回答的回调函数
		llms.WithJSONMode(), //返回json格式的数据
	)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(completion)

	//如需要设置System
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "你是一个地里老师"),
		llms.TextParts(llms.ChatMessageTypeHuman, "日本的首都是哪里?"),
	}
	response, err := llm.GenerateContent(ctx, content, llms.WithTemperature(0))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(response.Choices[0].Content)

	completion, err = llms.GenerateFromSinglePrompt(ctx,
		llm,
		"以上两个城市相距多少公里?",
		llms.WithTemperature(0.0),
		llms.WithJSONMode(),
	) //对话是无状态的，没有记住历史对话

	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(completion)
}

// conversation 会话，多轮问答
func conversation() {
	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		log.Fatalln(err)
	}
	//怎么指定Temperature？怎么指定System？怎么指定JSONMode？
	conversationBuffer := memory.NewConversationWindowBuffer(5) //记住前5次对话的内容
	llmChain := chains.NewConversation(llm, conversationBuffer)
	ctx := context.Background()
	completion, err := chains.Run(ctx, llmChain, "中国的首都是哪里?")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(completion)
	completion, err = chains.Run(ctx, llmChain, "日本的首都是哪里?")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(completion)
	completion, err = chains.Run(ctx, llmChain, "以上两个城市相距多少公里?")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(completion)
}

func main() {
	simpleQA()
	conversation()
}
