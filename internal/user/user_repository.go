package user

import (
	"context"
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

type DBTX interface {
}

type Repository struct {
	db *gocql.Session
}

func NewRepository(db *gocql.Session) REPOSITORY {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) (*User, error) {

	newId, _ := gocql.RandomUUID()

	if err := r.db.Query(`INSERT INTO users (id, username, email, name, password) 
                       VALUES (?, ?, ?, ?, ?)`,
		newId, user.Username, user.Email, user.Name, user.Password).WithContext(ctx).Exec(); err != nil {
		log.Fatal(err)
	}

	user.ID = newId
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := User{}
	scanner := r.db.Query(`SELECT id, username, email, name, password FROM users WHERE email=? ALLOW FILTERING`, email).WithContext(ctx).Iter().Scanner()
	scanner.Next()

	err := scanner.Scan(&u.ID, &u.Username, &u.Email, &u.Name, &u.Password)
	if err != nil {
		return nil, fmt.Errorf("Cannot scan: %v", err)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error while closing the scanner: %v", err)
	}

	return &u, nil
}
