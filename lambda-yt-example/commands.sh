#!/bin/bash

aws iam create-role --role-name lambda-ex --assume-role-policy-document '{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]}'
aws iam create-role --role-name lambda-ex --assume-role-policy-document file://trust-policy.json
aws iam attach-role-policy --role-name lambda-ex --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

go mod tidy
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go
zip function.zip main

aws lambda create-function --function-name go-lambda-function3 \
    --zip-file fileb://function.zip --handler main --runtime go1.x \
    --role arn:aws:iam::************:role/lambda-ex

aws lambda invoke --function-name go-lambda-function3 --cli-binary-format raw-in-base64-out --payload '{"What is your name?": "Jim", "How old are you?": 12}' output.txt
