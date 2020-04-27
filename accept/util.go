package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
)

func ChallengeHandler(request events.APIGatewayProxyRequest) (Response, error) {
	body := request.Body
	type ChallengeResponse struct {
		Challenge string
	}
	var r ChallengeResponse
	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		s := "Unable to parse challenge JSON!"
		return Response{
			Body:       s,
			StatusCode: 500,
		}, err
	}
	return Response{
		Body:       r.Challenge,
		StatusCode: 200,
	}, nil
}
