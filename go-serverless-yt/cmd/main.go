package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pirateunclejack/go-practice/go-serverless-yt/pkg/handlers"
)

const tableName = "go-serverless-yt"
var (
    dynaClient dynamodbiface.DynamoDBAPI
)

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
    switch req.HTTPMethod{
        case "GET":
            return handlers.GetUser(req, tableName, dynaClient)
        case "POST":
            return handlers.CreateUser(req, tableName, dynaClient)
        case "PUT":
            return handlers.UpdateUser(req, tableName, dynaClient)
        case "DELETE":
            return handlers.DeleteUser(req, tableName, dynaClient)
        default:
            return handlers.UnhandledMethod()
    }
}

func main() {
    region := os.Getenv("AWS_REGION")
    awsSession, err := session.NewSession(&aws.Config{
        Region: aws.String(region),})
    
    if err != nil {
        log.Println("faild to start aws session: ", err)
        return
    }

    dynaClient = dynamodb.New(awsSession)
    lambda.Start(handler)
}
