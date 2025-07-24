package main

import (
	"log"
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) error {
	if err := SeedUsers(db); err != nil {
		return err
	}
	
	if err := SeedProblems(db); err != nil {
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
	problems := []Problem{
		{
			Title:       "Two Sum",
			Description: "Given an array of integers nums and an integer target, return indices of the two numbers such that they add up to target.",
			Difficulty:  "Easy",
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