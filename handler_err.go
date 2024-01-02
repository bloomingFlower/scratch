package main

import (
	"fmt"
	api "github.com/bloomingFlower/rssagg/protos"
	"log"
)

func (s *server) handlerErr(req *api.ErrRequest) (*api.ErrResponse, error) {
	log.Printf("Received error request: %v", req.ErrorMessage)

	res := &api.ErrResponse{
		ResultMessage: fmt.Sprintf("Received error request: %v", req.ErrorMessage),
	}

	return res, nil
}
