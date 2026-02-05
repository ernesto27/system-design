package main

import (
	"encoding/json"
	"log"
	"gorm.io/gorm"
	"gorm.io/datatypes"
)

func SeedData(db *gorm.DB) error {
	if err := SeedUsers(db); err != nil {
		return err
	}
	
	if err := SeedProblems(db); err != nil {
		return err
	}
	
	if err := SeedCodeBases(db); err != nil {
		return err
	}
	
	return nil
}

func SeedUsers(db *gorm.DB) error {
	users := []User{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Smith", Email: "jane@example.com"},
		{Name: "Alice Johnson", Email: "alice@example.com"},
		{Name: "Bob Wilson", Email: "bob@example.com"},
		{Name: "Charlie Brown", Email: "charlie@example.com"},
	}
	
	for _, user := range users {
		var existingUser User
		if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&user).Error; err != nil {
				log.Printf("Failed to create user %s: %v", user.Name, err)
				return err
			}
			log.Printf("Created user: %s", user.Name)
		}
	}
	
	return nil
}

func SeedProblems(db *gorm.DB) error {
	twoSumTestCases := []map[string]interface{}{
		{
			"name":            "Basic Input",
			"nums":            []int{2, 7, 11, 15},
			"target":          9,
			"expected_output": []int{0, 1},
		},
		{
			"name":            "Negative Numbers",
			"nums":            []int{3, -4, 5, 8},
			"target":          1,
			"expected_output": []int{1, 2},
		},
		{
			"name":            "Duplicates in Array",
			"nums":            []int{3, 3},
			"target":          6,
			"expected_output": []int{0, 1},
		},
	}
	
	twoSumTestCasesJSON, _ := json.Marshal(twoSumTestCases)
	
	problems := []Problem{
		{
			Title:       "Two Sum",
			Description: "Given an array of integers nums and an integer target, return indices of the two numbers such that they add up to target.",
			Difficulty:  "Easy",
			TestCases:   datatypes.JSON(twoSumTestCasesJSON),
		},
		{
			Title:       "Add Two Numbers",
			Description: "You are given two non-empty linked lists representing two non-negative integers.",
			Difficulty:  "Medium",
		},
		{
			Title:       "Longest Substring Without Repeating Characters",
			Description: "Given a string s, find the length of the longest substring without repeating characters.",
			Difficulty:  "Medium",
		},
		{
			Title:       "Median of Two Sorted Arrays",
			Description: "Given two sorted arrays nums1 and nums2 of size m and n respectively, return the median of the two sorted arrays.",
			Difficulty:  "Hard",
		},
		{
			Title:       "Longest Palindromic Substring",
			Description: "Given a string s, return the longest palindromic substring in s.",
			Difficulty:  "Medium",
		},
	}
	
	for _, problem := range problems {
		var existingProblem Problem
		if err := db.Where("title = ?", problem.Title).First(&existingProblem).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&problem).Error; err != nil {
				log.Printf("Failed to create problem %s: %v", problem.Title, err)
				return err
			}
			log.Printf("Created problem: %s", problem.Title)
		}
	}
	
	return nil
}

func SeedCodeBases(db *gorm.DB) error {
	// Get Two Sum problem ID
	var twoSumProblem Problem
	if err := db.Where("title = ?", "Two Sum").First(&twoSumProblem).Error; err != nil {
		log.Printf("Two Sum problem not found, skipping code base seeding: %v", err)
		return nil
	}

	codeBases := []CodeBase{
		{
			ProblemID: twoSumProblem.ID,
			Language:  "javascript",
			Template: `/**
 * @param {number[]} nums
 * @param {number} target
 * @return {number[]}
 */
var twoSum = function(nums, target) {
    
};`,
		},
		{
			ProblemID: twoSumProblem.ID,
			Language:  "python",
			Template: `def two_sum(nums, target):
    """
    :type nums: List[int]
    :type target: int
    :rtype: List[int]
    """
    pass`,
		},
		{
			ProblemID: twoSumProblem.ID,
			Language:  "go",
			Template: `func twoSum(nums []int, target int) []int {
    
}`,
		},
	}

	for _, codeBase := range codeBases {
		var existingCodeBase CodeBase
		if err := db.Where("problem_id = ? AND language = ?", codeBase.ProblemID, codeBase.Language).First(&existingCodeBase).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&codeBase).Error; err != nil {
				log.Printf("Failed to create code base for problem %d, language %s: %v", codeBase.ProblemID, codeBase.Language, err)
				return err
			}
			log.Printf("Created code base: Problem %d, Language %s", codeBase.ProblemID, codeBase.Language)
		}
	}

	return nil
}