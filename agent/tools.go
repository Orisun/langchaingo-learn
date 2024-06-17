package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type StudentScoreTool struct{}

func (StudentScoreTool) Name() string {
	return "student_score"
}

// 工具的描述很关键，告诉大模型什么情况下使用这个工具，以及工具的输入是什么
func (StudentScoreTool) Description() string {
	return "use this tool when you need get the score of a student, please input the student's name"
}
func (StudentScoreTool) Call(ctx context.Context, input string) (string, error) {
	name := strings.TrimSuffix(input, "'s name")
	fmt.Println(input, name)
	if score, err := GetScoreOfStudent(name); err != nil {
		return "", err
	} else {
		return strconv.FormatFloat(score, 'f', 2, 64), nil
	}
}

type StudentCityTool struct{}

func (StudentCityTool) Name() string {
	return "student_city"
}

// 工具的描述很关键，告诉大模型什么情况下使用这个工具，以及工具的输入是什么
func (StudentCityTool) Description() string {
	return "use this tool when you need get the city of a student, please input the student's name"
}
func (StudentCityTool) Call(ctx context.Context, input string) (string, error) {
	name := strings.TrimSuffix(input, "'s name")
	fmt.Println(input, name)
	return GetCityOfStudent(name)
}
