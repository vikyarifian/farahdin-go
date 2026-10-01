// Command import-convex copies the source app's Convex data into PostgreSQL.
//
//	npx convex export --path convex-export.zip   (in farahdin-react-native)
//	POSTGRES_URL=postgres://... go run ./cmd/import-convex convex-export.zip
//
// Imported users keep their email; on their first Google sign-in they are
// linked by email, exactly like the Clerk webhook's createUser did.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/vikyarifian/farahdin-go/internal/repository"
)

func main() {
	dbURL := flag.String("db", os.Getenv("POSTGRES_URL"), "PostgreSQL URL (default $POSTGRES_URL)")
	flag.Parse()
	if flag.NArg() != 1 || *dbURL == "" {
		fmt.Fprintln(os.Stderr, "usage: POSTGRES_URL=postgres://... import-convex convex-export.zip")
		os.Exit(2)
	}
	ctx := context.Background()
	db, err := repository.Open(*dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	if err := repository.Migrate(ctx, db); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	st, err := repository.ImportConvex(ctx, db, flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "import failed (nothing was written):", err)
		os.Exit(1)
	}
	fmt.Printf("imported users=%d categories=%d inboxes=%d\n", st.Users, st.Categories, st.Inboxes)
}
