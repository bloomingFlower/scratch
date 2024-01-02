package main

import (
	"bytes"
	"context"
	api "github.com/bloomingFlower/rssagg/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"text/template"
)

func (s *server) View(ctx context.Context, req *api.ViewRequest) (*api.ViewResponse, error) {
	// HTML 파일을 파싱합니다.
	htmlTemplate, err := template.ParseFiles("html/view.html")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error generate view: %v", err)
	}

	// 템플릿을 실행하고 출력을 문자열로 작성합니다.
	var htmlOutput bytes.Buffer
	err = htmlTemplate.Execute(&htmlOutput, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error generate view: %v", err)
	}

	// 응답을 생성합니다.
	res := &api.ViewResponse{
		Html: htmlOutput.String(),
	}

	return res, nil
}
