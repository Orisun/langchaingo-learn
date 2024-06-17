/**
RAG，即检索增强生成 (Retrieval-Augmented Generation)，实际上是将知识检索 (Retrieval) 和语言生成 (Generation) 两种技术巧妙地结合在一起。它的核心思想是，在生成回答或文本时，先从海量的文档知识库中检索出与问题最相关的几段文本，然后以此为基础再衍生出连贯自然的回答。
**/

package main

import (
	"context"
	"fmt"
	"langchaingo-learn/util/segger"
	"log"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

var (
	text = []string{
		"话说唐僧师徒四人西天取经，途经白虎岭。",
		"白虎岭上有个妖怪，名叫白骨精。",
		"她为了吃唐僧肉，就变幻成一个美丽的女子，来引诱唐僧。",
		"刘备留下关张二人,以兄事之。",
		"三人各自对天地、日月星辰发誓,然后两两互持桃枝彼此磕头,作兄弟之礼,场面极为隆重感人。",
		"次日,三人同至桃园,刘备事先示意关张二人,各持一枝桃花,致告天地曰:",
		"武大郎忍痛抽筋,往后瞧时,只见那猛虎抡起铁棒,向自己直扑将来。",
		"武大郎使开双戟,侧身让过。那猛虎扑了一空,回头又扑。",
		"武大郎顺手又一刀,把它一只后腿也砍断。",
	}
)

// 存储每段文本对应的向量
func StoreEmbedding(text []string, llm embeddings.EmbedderClient) {
	embeder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatalln(err)
	}

	//pip install chromadb
	//chroma run --host localhost --port 8000 --path ./my_chroma_data
	//pip install pydantic --upgrade

	store, err := chroma.New( //支持数十种向量数据库，chroma非常的轻量级，且可以完全存放在内存中。milvus和weaviate都有go api
		chroma.WithChromaURL("http://localhost:8000"),
		chroma.WithEmbedder(embeder),
		chroma.WithDistanceFunction("cosine"),
		// chroma.WithNameSpace(uuid.New().String()),
		chroma.WithNameSpace("test"),
	)
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	docs := make([]schema.Document, 0, len(text))
	for i, s := range text {
		doc := schema.Document{PageContent: strings.Join(segger.Cut(s), " "), Metadata: map[string]any{"id": i}, Score: 0.0}
		docs = append(docs, doc)
		if len(docs) >= 3 { //量太大的话，调LLM获取embedding时会超时
			_, err = store.AddDocuments(ctx, docs)
			if err != nil {
				log.Fatalln(err)
			}
			docs = make([]schema.Document, 0, len(text))
		}
	}
}

// 寻找跟query相关的文本
func GetRelevantDocuments(query string, llm embeddings.EmbedderClient) []string {
	embeder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatalln(err)
	}

	store, err := chroma.New( //支持数十种向量数据库，chroma非常的轻量级，且可以完全存放在内存中。milvus和weaviate都有go api
		chroma.WithChromaURL("http://localhost:8000"),
		chroma.WithEmbedder(embeder),
		chroma.WithDistanceFunction("cosine"),
		// chroma.WithNameSpace(uuid.New().String()),
		chroma.WithNameSpace("test"),
	)
	if err != nil {
		log.Fatalln(err)
	}

	var neighbors []schema.Document
	searchQuery := strings.Join(segger.Cut(query), " ")
	optionsVector := []vectorstores.Option{vectorstores.WithScoreThreshold(0.0)}
	ctx := context.Background()
	// 寻找最相似的邻居(方式一)
	neighbors, err = store.SimilaritySearch(ctx, searchQuery, 3, optionsVector...) //相似度位于[0,1]上
	// 寻找最相似的邻居(方式二)
	// retriever := vectorstores.ToRetriever(store, 3, optionsVector...)
	// neighbors, err = retriever.GetRelevantDocuments(ctx, searchQuery)

	if err != nil {
		log.Fatalln(err)
	}

	rect := make([]string, 0, len(neighbors))
	for _, neighbor := range neighbors {
		if index, ok := neighbor.Metadata["id"].(float64); ok {
			i := int(index)
			// fmt.Printf("%s\n", text[i])
			rect = append(rect, text[i])
		} else {
			fmt.Printf("%T\n", neighbor.Metadata["id"])
		}
	}

	return rect
}

func main() {
	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		log.Fatalln(err)
	}

	// StoreEmbedding(text, llm)

	query := "武大郎在打什么？"
	relevantText := GetRelevantDocuments(query, llm)
	prompt := strings.Join(relevantText, "\n") + "\n请根据以上背景知识，回答这个问题：" + query
	fmt.Println(prompt)
	// 用相关文本作为背景知识提供给LLM，再让LLM回答这个问题
	completion, err := llms.GenerateFromSinglePrompt(
		context.Background(),
		llm,
		prompt,
		llms.WithTemperature(0.0),
	)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(completion)
}
