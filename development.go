// +build development

package main

import "fmt"

func init() {
	// set development flag to true, which uses the tpl/home.html file
	// rather than imbedding it
	InDevelopment = true
}
