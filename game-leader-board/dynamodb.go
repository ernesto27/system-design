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
	client    *dynamodb.Client
	tableName string
}

type ScoreRequest struct {
	UserID int    `json:"user_id"`
	GameID string `json:"game_id"`
	Name   string `json:"name"`
}

type LeaderBoard struct {
	GameID string `json:"game_id"`
	UserID int    `json:"user_id"`
	Score  int    `json:"score" `
	Rank   int    `json:"rank"`
}

type LeaderBoardByUserID struct {
	LeaderBoard
}

func NewDynamoDB(key string, secret string, region string, tableName string) *DynamoDB {
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
		client:    ddbClient,
		tableName: tableName,
	}
}

func (d *DynamoDB) PutItem(scoreRequest ScoreRequest) error {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"game_id": &types.AttributeValueMemberS{Value: scoreRequest.GameID},
			"user_id": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", scoreRequest.UserID)},
		},
	}

	result, err := d.client.GetItem(context.TODO(), input)
	if err != nil {
		return err
	}

	var scoreValue int
	if result.Item == nil {
		scoreValue = 1
	} else {
		scoreValue, err = strconv.Atoi(result.Item["score"].(*types.AttributeValueMemberN).Value)
		if err != nil {
			return err
		}
		scoreValue += 1
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item: map[string]types.AttributeValue{
			"game_id": &types.AttributeValueMemberS{Value: scoreRequest.GameID},
			"user_id": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", scoreRequest.UserID)},
			"score":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", scoreValue)},
		},
	}

	_, err = d.client.PutItem(context.TODO(), putItemInput)
	if err != nil {
		return err
	}

	return nil
}

func (d *DynamoDB) GetLeaderboard(gameID string) ([]LeaderBoard, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		IndexName:              aws.String("game_id-score-index"),
		KeyConditionExpression: aws.String("game_id = :ln"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":ln": &types.AttributeValueMemberS{Value: gameID},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(10),
	}

	result, err := d.client.Query(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var leaderBoard []LeaderBoard
	rank := 1
	for _, item := range result.Items {
		scoreValue, err := strconv.Atoi(item["score"].(*types.AttributeValueMemberN).Value)
		if err != nil {
			return nil, err
		}
		userIDValue, err := strconv.Atoi(item["user_id"].(*types.AttributeValueMemberN).Value)
		if err != nil {
			return nil, err
		}

		leaderBoard = append(leaderBoard, LeaderBoard{
			GameID: item["game_id"].(*types.AttributeValueMemberS).Value,
			UserID: userIDValue,
			Score:  scoreValue,
			Rank:   rank,
		})

		rank++
	}

	return leaderBoard, nil
}

func (d *DynamoDB) GetLeaderBoardByUserID(scoreRequest ScoreRequest) (LeaderBoardByUserID, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"game_id": &types.AttributeValueMemberS{Value: scoreRequest.GameID},
			"user_id": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", scoreRequest.UserID)},
		},
	}

	var leaderBoard LeaderBoardByUserID
	result, err := d.client.GetItem(context.TODO(), input)
	if err != nil {
		return leaderBoard, err
	}

	if result.Item == nil {
		return leaderBoard, fmt.Errorf("user_id not found")
	}

	scoreValue, err := strconv.Atoi(result.Item["score"].(*types.AttributeValueMemberN).Value)
	if err != nil {
		return leaderBoard, err
	}

	gameId := scoreRequest.GameID
	userScore := scoreValue

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		IndexName:              aws.String("game_id-score-index"),
		KeyConditionExpression: aws.String("game_id = :game_id AND score > :score"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":game_id": &types.AttributeValueMemberS{Value: gameId},
			":score":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", userScore)},
		},
		Select: types.SelectAllAttributes,
	}

	queryResult, err := d.client.Query(context.TODO(), queryInput)
	if err != nil {
		return leaderBoard, err
	}

	fmt.Println(queryResult.Items)

	leaderBoard.GameID = scoreRequest.GameID
	leaderBoard.UserID = scoreRequest.UserID
	leaderBoard.Score = scoreValue
	leaderBoard.Rank = int(queryResult.Count) + 1

	return leaderBoard, nil
}

func (d *DynamoDB) Test() {
	// Get ranking
	// input := &dynamodb.QueryInput{
	// 	TableName:              aws.String("test_leaderboar"),
	// 	IndexName:              aws.String("game_id-score-index"),
	// 	KeyConditionExpression: aws.String("game_id = :ln"),
	// 	ExpressionAttributeValues: map[string]types.AttributeValue{
	// 		":ln": &types.AttributeValueMemberS{Value: "game1"},
	// 	},
	// 	ScanIndexForward: aws.Bool(true),
	// }

	// result, err := d.client.Query(context.TODO(), input)
	// if err != nil {
	// 	panic(err)
	// }

	// for _, item := range result.Items {
	// 	fmt.Println(item["score"].(*types.AttributeValueMemberN).Value)
	// }

	// Get by user by id
	// input := &dynamodb.GetItemInput{
	// 	TableName: aws.String("test_leaderboar"),
	// 	Key: map[string]types.AttributeValue{
	// 		"game_id": &types.AttributeValueMemberS{Value: "game1"},
	// 		"user_id": &types.AttributeValueMemberN{Value: "12"},
	// 	},
	// }

	// result, err := d.client.GetItem(context.TODO(), input)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(result.Item["user_id"].(*types.AttributeValueMemberN).Value)

	// Put item
	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String("test_leaderboar"),
		Item: map[string]types.AttributeValue{
			"game_id": &types.AttributeValueMemberS{Value: "game1"},
			"user_id": &types.AttributeValueMemberN{Value: "100"},
			"score":   &types.AttributeValueMemberN{Value: "66"},
		},
	}

	_, err := d.client.PutItem(context.TODO(), putItemInput)
	if err != nil {
		panic(err)
	}

	fmt.Println("Put item success")

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
