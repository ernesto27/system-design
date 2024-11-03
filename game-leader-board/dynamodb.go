package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDB struct {
	client *dynamodb.Client
}

type ScoreRequest struct {
	UserID       int    `json:"user_id"`
	GameID       string `json:"game_id"`
	CurrentScore string `json:"current_score"`
	Name         string `json:"name"`
}

type LeaderBoard struct {
	LeaderboardName string `json:"leaderboard_name"`
	ScoreSort       string `json:"-"`
	Score           string `json:"score"`
	Name            string `json:"name"`
}

func NewDynamoDB(key string, secret string, region string) *DynamoDB {
	creds := credentials.NewStaticCredentialsProvider(key, secret, "")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		log.Fatal(err)
	}

	ddbClient = dynamodb.NewFromConfig(cfg)
	return &DynamoDB{
		client: ddbClient,
	}
}

func (d *DynamoDB) PutItem(scoreRequest ScoreRequest) error {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("leaderboard"),
		Key: map[string]types.AttributeValue{
			"leaderboard_name": &types.AttributeValueMemberS{Value: scoreRequest.GameID},
			"score":            &types.AttributeValueMemberS{Value: scoreRequest.CurrentScore},
		},
	}

	result, err := myDB.client.GetItem(context.TODO(), input)
	if err != nil {
		return err
	}

	if result.Item != nil {
		fmt.Println("delete current item")
		input := &dynamodb.DeleteItemInput{
			TableName: aws.String("leaderboard"),
			Key: map[string]types.AttributeValue{
				"leaderboard_name": &types.AttributeValueMemberS{Value: scoreRequest.GameID},
				"score":            &types.AttributeValueMemberS{Value: scoreRequest.CurrentScore},
			},
		}

		_, err := ddbClient.DeleteItem(context.TODO(), input)

		if err != nil {
			return err
		}
	}

	var currScore int
	if scoreRequest.CurrentScore == "666" {
		currScore = 1
	} else {
		currScore, err = parseScoreFromString(scoreRequest.CurrentScore)
		if err != nil {
			return err
		}
		currScore++
	}

	newScore := padScore(currScore, scoreRequest.UserID)

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String("leaderboard"),
		Item: map[string]types.AttributeValue{
			"leaderboard_name": &types.AttributeValueMemberS{Value: scoreRequest.GameID},
			"score":            &types.AttributeValueMemberS{Value: newScore},
			"score_value":      &types.AttributeValueMemberN{Value: strconv.Itoa(currScore)},
			"name":             &types.AttributeValueMemberS{Value: scoreRequest.Name},
		},
	}

	_, err = d.client.PutItem(context.TODO(), putItemInput)

	return err
}

func (d *DynamoDB) GetLeaderboard(leaderboardName string) ([]LeaderBoard, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String("leaderboard"),
		KeyConditionExpression: aws.String("leaderboard_name = :ln"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":ln": &types.AttributeValueMemberS{Value: leaderboardName},
		},
		ScanIndexForward: aws.Bool(false),
	}

	result, err := d.client.Query(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var leaderBoard []LeaderBoard
	for _, item := range result.Items {
		leaderBoard = append(leaderBoard, LeaderBoard{
			LeaderboardName: item["leaderboard_name"].(*types.AttributeValueMemberS).Value,
			Score:           item["score_value"].(*types.AttributeValueMemberN).Value,
			Name:            item["name"].(*types.AttributeValueMemberS).Value,
		})

	}

	return leaderBoard, nil
}

func padScore(score int, userID int) string {
	return fmt.Sprintf("%010d#%d", score, userID)
}

func parseScoreFromString(scoreUserId string) (int, error) {
	parts := strings.Split(scoreUserId, "#")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid score_userId format")
	}

	score, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("failed to parse score: %v", err)
	}

	return score, nil
}
