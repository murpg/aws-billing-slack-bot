package main

import (
	"context"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/now"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

var webhookUrl = os.Getenv("SLACK_WEBHOOK")
var c redis.Conn

func init() {
	c, _ = redis.Dial("tcp", os.Getenv("REDIS_HOST"))
	c.Do("AUTH", os.Getenv("REDIS_AUTH"))
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	granularity := "MONTHLY"
	metrics := []string{"UnblendedCost",}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create Cost Explorer Service Client
	svc := costexplorer.New(sess)
	result, err := svc.GetCostAndUsage(&costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(now.BeginningOfMonth().Format("2006-01-02")),
			End:   aws.String(now.EndOfMonth().Format("2006-01-02")),
		},
		Granularity: aws.String(granularity),
		GroupBy: []*costexplorer.GroupDefinition{
			{
				Type: aws.String("DIMENSION"),
				Key:  aws.String("SERVICE"),
			},
		},
		Metrics: aws.StringSlice(metrics),
	})
	if err != nil {
		return Response{Body: "Error", StatusCode: 500}, err
	}
	for _, group := range result.ResultsByTime[0].Groups {
		r, err := redis.Float64(c.Do("GET", utilizeString(group.Keys[0])))
		if err == redis.ErrNil {
			continue
		}
		if r < convertF(group.Metrics["UnblendedCost"].Amount) {
			slack.Send(webhookUrl, "", slack.Payload{Text: ":warning:\nExceeded Threshold: " + *group.Keys[0] + "\nCost: " + *group.Metrics["UnblendedCost"].Amount})
		}
	}
	return Response{Body: "OK", StatusCode: 200,}, nil
}

func main() {
	lambda.Start(Handler)
}
