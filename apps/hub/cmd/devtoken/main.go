// Command devtoken prints a signed JWT for local API testing.
// Usage: devtoken <user-id> <org-id> (JWT_SECRET from env)
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/harmost/hub/internal/auth"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: devtoken <user-id> <org-id>")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	token, err := auth.Sign(os.Args[1], os.Args[2], secret)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(token)
}
