//go:build development
// +build development

package web

func init() {
	// set development flag to true, which uses the templates/home.html file
	// rather than imbedding it
	inDevelopment = true
}
