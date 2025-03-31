package main

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	Name   string
	Author string
	Page   uint
}

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string
}
