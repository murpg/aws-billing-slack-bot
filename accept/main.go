package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/gomodule/redigo/redis"
	"github.com/slack-go/slack/slackevents"
	"os"
	"strings"

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


func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (Response, error) {
	var cmd, service, threshold, tmp string
	if strings.Contains(req.Body, "challenge") {
		return ChallengeHandler(req)
	}

	// Read the command
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(req.Body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: os.Getenv("SLACK_VERIFICATION_TOKEN")}))
	if err != nil {
		return Response{Body: "error", StatusCode: 500}, err
	}
	event := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)

	// ADD COMMAND
	if strings.Contains(event.Text, "ADD") {
		fmt.Sscanf(event.Text, "%s %s %s %s", &tmp, &cmd, &service, &threshold)
		if _, err = c.Do("SET", service, threshold); err == nil {
			slack.Send(webhookUrl, "", slack.Payload{Text: "Alert added for service: " + service})
		}
	}

	// LIST COMMAND
	if strings.Contains(event.Text, "LIST") {
		return listAllAlerts()
	}
	return Response{Body: "OK", StatusCode: 200,}, nil
}

func listAllAlerts() (Response, error) {
	keys, _ := redis.Strings(c.Do("KEYS", "*"))
	attachment := slack.Attachment{}
	for _, key := range keys {
		r, _ := redis.String(c.Do("GET", key))
		attachment.AddField(slack.Field{Title: key + " " + r + "$"})
	}
	attachment.AddAction(slack.Action{Type: "button", Text: "Aws Console :chart:", Url: "https://console.aws.amazon.com/console/home", Style: "primary"})
	payload := slack.Payload{
		Text:        "Threshold Alert",
		Username:    "AWS Billing Bot",
		IconEmoji:   ":monkey_face:",
		Attachments: []slack.Attachment{attachment},
	}
	errs := slack.Send(webhookUrl, "", payload)
	if len(errs) > 0 {
		return Response{Body: "OK", StatusCode: 200,}, errs[0]
	}
	return Response{Body: "OK", StatusCode: 200,}, nil
}

func main() {
	lambda.Start(Handler)
}
