package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rilldata/rill/server-cloud/ent"
	"github.com/rilldata/rill/server-cloud/ent/organization"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	ctx := context.Background()
	// Run the auto migration tool.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	if _, err = CreateOrganization(ctx, client); err != nil {
		log.Fatal(err)
	}
	if _, err = QueryOrganization(ctx, client); err != nil {
		log.Fatal(err)
	}
}

func CreateOrganization(ctx context.Context, client *ent.Client) (*ent.Organization, error) {
	u, err := client.Organization.
		Create().
		SetName("Test").
		SetDescription("This is first Organization").
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed creating Organization: %w", err)
	}
	log.Println("Organization was created: ", u)
	return u, nil
}

func QueryOrganization(ctx context.Context, client *ent.Client) (*ent.Organization, error) {
	u, err := client.Organization.
		Query().
		Where(organization.NameEQ("Test")).
		// `Only` fails if no user found,
		// or more than 1 user returned.
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying Organization: %w", err)
	}
	log.Println("Organization returned: ", u)
	return u, nil
}
