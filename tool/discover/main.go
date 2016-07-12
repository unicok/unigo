package main

import "os"

const (
	raw = "https://raw.githubusercontent.com/amorwilliams/bodoni/master/lib/services/services.go"
)

func main() {
	if len(os.Args) <= 1 {
		return
	}
	// resp, err := http.Get(raw)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
