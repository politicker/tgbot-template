package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

var queueName = "telegram-bot-template"

func (app *appEnv) Send(ctx context.Context, untaggedLogger *zap.Logger) error {
	logger := untaggedLogger.With(zap.String("queue-name", queueName))
	logger.Info("beginning queue loop")

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sqs.New(sess)

	accountID := os.Getenv("SQS_ACCOUNT_ID")
	if accountID == "" {
		log.Fatal("must specify SQS_ACCOUNT_ID env")
	}

	queueURL := fmt.Sprintf("https://sqs.us-east-1.amazonaws.com/%s/%s", accountID, queueName)

	for {
		msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: aws.Int64(1),
			VisibilityTimeout:   aws.Int64(60),
			WaitTimeSeconds:     aws.Int64(20),
		})
		if err != nil {
			logger.Error("failed to receive messages", zap.Error(err))
			time.Sleep(60 * time.Second)
			continue
		}

		for _, message := range msgResult.Messages {
			logger.Info("received on sqs")

			// TODO: We need to know the shape of the object here for deserialization
			var obj T
			err := json.Unmarshal([]byte(*message.Body), &obj)
			if err != nil {
				logger.Info(*message.Body)
				logger.Error("failed to parse json", zap.Error(err))
				time.Sleep(60 * time.Second)
				continue
			}

			// I don't hate having the handler be a passed in function or even something
			// defined in this file. The business logic of handling a message probably shouldn't be
			// inline here. Also, is there anything more to handling a queue event than sending a telegram
			// message? We probably want to keep it wrapped with a handler to deal with any boilerplate init/error
			// handling stuff.
			err = app.sendMessage(ctx, obj)
			err = handler(ctx, logger, &obj)
			if err != nil {
				logger.Error("failed to handle message", zap.Error(err))
				time.Sleep(60 * time.Second)
				continue
			}

			_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				logger.Error("failed to delete message", zap.Error(err))
				time.Sleep(60 * time.Second)
				continue
			}
		}
	}
}

// TODO: We need to define an SQS payload type that we can pull messageBody data from
func (app *appEnv) sendMessage(ctx context.Context, messageBody any) error {
	cfg := telegram.NewMessage(messageBody.GroupID, messageBody.Message)
	_, err := app.bot.Send(cfg)
	if err != nil {
		app.logger.Error("failed to send message to chat", zap.Error(err), zap.Any("message", messageBody))
	}

	return nil
}
