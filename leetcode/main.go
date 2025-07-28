package main

import (
	codeexecutor "leetcode/internal/code_executor"
	"log"
	"os"
)

func main() {
	testCases := []codeexecutor.TestCaseData{
		{
			Name: "Basic Input",
			Parameters: map[string]interface{}{
				"nums":   []int{2, 7, 11, 15},
				"target": 9,
			},
			ExpectedOutput: []int{0, 1},
		},
		// {
		// 	Name: "Negative Numbers",
		// 	Parameters: map[string]interface{}{
		// 		"nums":   []int{-3, 4, 3, 90},
		// 		"target": 0,
		// 	},
		// 	ExpectedOutput: []int{1, 2},
		// },
		// {
		// 	Name: "Duplicates in Array",
		// 	Parameters: map[string]interface{}{
		// 		"nums":   []int{3, 3, 4},
		// 		"target": 6,
		// 	},
		// 	ExpectedOutput: []int{0, 1},
		// },
	}

	code := "function twoSum(nums, target) { return [9, 1]; };  console.log(twoSum($1, $2));"

	codeexecutor.RunAllTestCases("js", code, testCases)
	os.Exit(1)

	config := NewDatabaseConfig()

	db, err := ConnectDatabase(config)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if err := SeedData(db); err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	log.Println("Database connected, migrated, and seeded successfully!")

	router := SetupRouter(db)
	log.Println("Starting server on :8080")
	router.Run(":8080")
}
