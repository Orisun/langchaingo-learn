package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi"
)

func main() {
	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		log.Fatalln(err)
	}

	search, err := serpapi.New() //去https://serpapi.com/上注册一个账号，获取key，放到环境变量SERPAPI_API_KEY里，serpapi支持搜索google、bing、Baidu、Yelp等等。目前go版本只支持google
	if err != nil {
		log.Fatalln(err)
	}

	agentTools := []tools.Tool{
		tools.Calculator{}, //LangChain内置的Tool，用于数学计算
		search,             //如果包含了search tool，LLM查询学生成绩时会去询问Google
		StudentScoreTool{}, //自定义的Tool
		StudentCityTool{},  //自定义的Tool
	}
	executor, err := agents.Initialize(
		llm,
		agentTools,
		agents.ZeroShotReactDescription, //ZeroShotReactDescription--一次对话，ConversationalReactDescription--多轮对话
		agents.WithMaxIterations(10),    //经过几轮得到答案，每一轮都要经历:
		// Thought--大模型需要思考一下是否需要使用工具，以及使用哪个工具
		// Action--决定使用哪个工具
		// Action Input--准备好工具的输入
		// Observation--得到工具的输出结果
	)
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	var answer string

	answer, err = chains.Run(ctx, executor, "Tom住在哪个城市", chains.WithTemperature(0.3))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(answer)

	answer, err = chains.Run(ctx, executor, "Tom住的城市离上海有多远", chains.WithTemperature(0.3))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(answer)

	answer, err = chains.Run(ctx, executor, "Tom的成绩及格了吗", chains.WithTemperature(0.3))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(answer)
}
