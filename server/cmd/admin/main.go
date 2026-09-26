// Command admin manages users from the server itself. It talks to the
// database directly, so it works before anyone can log in (first admin)
// and when nobody can (password recovery).
//
//	go run ./cmd/admin create-admin --email you@example.com --name "Your Name"
//	go run ./cmd/admin reset-password --email you@example.com
package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sudarshanpokhrell/air/internal/config"
	"github.com/sudarshanpokhrell/air/internal/db"
	"github.com/sudarshanpokhrell/air/internal/store"
	"github.com/sudarshanpokhrell/air/internal/validator"
)

const usage = `Usage: go run ./cmd/admin <command> [flags]

Commands:
  create-admin   --email <email> --name <name>   Create a global admin and print its password
  reset-password --email <email>                 Set a new random password and print it`

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal(usage)
	}

	switch os.Args[1] {
	case "create-admin":
		createAdmin(os.Args[2:])
	case "reset-password":
		resetPassword(os.Args[2:])
	default:
		log.Fatal(usage)
	}
}

func openStore() (store.Store, func()) {
	cfg := config.MustLoad()

	database, err := db.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return store.NewStore(database), func() { database.Close() }
}

func createAdmin(args []string) {
	fs := flag.NewFlagSet("create-admin", flag.ExitOnError)
	email := fs.String("email", "", "email of the new admin")
	name := fs.String("name", "", "display name of the new admin")
	fs.Parse(args)

	user := &store.User{
		Email:   strings.TrimSpace(*email),
		Name:    strings.TrimSpace(*name),
		IsAdmin: true,
	}

	// The password is generated, never typed on the command line,
	// so it doesn't end up in shell history.
	password := rand.Text()

	v := validator.New()
	if store.ValidateUser(v, user, password); !v.Valid() {
		for field, msg := range v.Errors {
			log.Printf("--%s %s", field, msg)
		}
		os.Exit(1)
	}

	if err := user.Password.Set(password); err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	s, closeDB := openStore()
	defer closeDB()

	if err := s.Users.CreateUser(context.Background(), user); err != nil {
		if errors.Is(err, store.ErrConflict) {
			log.Fatalf("a user with email %q already exists (use reset-password)", user.Email)
		}
		log.Fatalf("failed to create user: %v", err)
	}

	fmt.Printf("Created global admin %s <%s>\n", user.Name, user.Email)
	fmt.Printf("Password: %s\n", password)
	fmt.Println("Log in and change it with POST /api/v1/me/password.")
}

func resetPassword(args []string) {
	fs := flag.NewFlagSet("reset-password", flag.ExitOnError)
	email := fs.String("email", "", "email of the user whose password to reset")
	fs.Parse(args)

	if strings.TrimSpace(*email) == "" {
		log.Fatal("--email is required")
	}

	s, closeDB := openStore()
	defer closeDB()

	ctx := context.Background()

	user, err := s.Users.GetUserByEmail(ctx, strings.TrimSpace(*email))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			log.Fatalf("no user with email %q", *email)
		}
		log.Fatalf("failed to load user: %v", err)
	}

	password := rand.Text()
	if err := user.Password.Set(password); err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}
	if err := s.Users.UpdatePassword(ctx, user); err != nil {
		log.Fatalf("failed to save password: %v", err)
	}

	// Anyone logged in with the old password is logged out.
	if err := s.Sessions.DeleteUserSessions(ctx, user.ID); err != nil {
		log.Fatalf("failed to delete sessions: %v", err)
	}

	fmt.Printf("New password for %s: %s\n", user.Email, password)
	if !user.IsActive {
		fmt.Println("Note: this account is deactivated and can't log in until an admin reactivates it.")
	}
}
