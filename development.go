// +build development

package main

import "fmt"

func init() {
	fmt.Println("using InDevelopment = true for development")
	InDevelopment = true
}
