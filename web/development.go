//go:build development
// +build development

package web

import "fmt"

func init() {
	// set development flag to true, which uses the tpl/home.html file
	// rather than imbedding it
	InDevelopment = true
}
