package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/config"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/repository/adapter"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/repository/instance"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/routes"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/rules"
	RulesProduct "github.com/pirateunclejack/go-practice/dynamodb-go-crud/internal/rules/product"
	"github.com/pirateunclejack/go-practice/dynamodb-go-crud/utils/logger"
)

func callMigrateAndAppendError(errors *[]error, connection *dynamodb.DynamoDB, rule rules.Interface)  {
    err := rule.Migrate(connection)
    if err != nil {
        *errors = append(*errors, err)
    }
}

func Migrate(connection *dynamodb.DynamoDB) []error {
	var errors []error
    callMigrateAndAppendError(&errors, connection, &RulesProduct.Rules{})

	return errors
}

func checkTables(connection *dynamodb.DynamoDB) error {
    response, err := connection.ListTables(&dynamodb.ListTablesInput{})
    if response != nil {
        if len(response.TableNames) == 0 {
            logger.INFO("tables not found", nil)
        }
    }

    for _, tableName := range response.TableNames {
        logger.INFO("table found: ", *tableName)
    }

    return err
}

func main() {
	configs := config.GetConfig()
	connection := instance.GetConnection()
	repository := adapter.NewAdapter(connection)
	logger.INFO("waiting for the service to start...", nil)

	errors := Migrate(connection)
	if len(errors) > 0 {
		for _, err := range errors {
			logger.PANIC("error on migration:...", err)
		}
	}

	logger.PANIC("", checkTables(connection))

	port := fmt.Sprintf(":%v", configs.Port)
	router := routes.NewRouter().SetRouters(repository)
	logger.INFO("service is running on port: ", port)

	server := http.ListenAndServe(port, router)
	log.Fatal(server)
}