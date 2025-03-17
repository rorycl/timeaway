module github.com/rorycl/timeaway

go 1.24

replace github.com/rorycl/timeaway/trips => ./trips

replace github.com/rorycl/timeaway/web => ./web

replace github.com/rorycl/timeaway/cmd => ./cmd

require (
	github.com/ajstarks/svgo v0.0.0-20211024235047-1546f124cd8b
	github.com/go-playground/form v3.1.4+incompatible
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/jessevdk/go-flags v1.6.1
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)
