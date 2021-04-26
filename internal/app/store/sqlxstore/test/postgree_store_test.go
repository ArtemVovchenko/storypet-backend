package sqlxstore

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/sqlxstore"
	"log"
	"os"
	"testing"
)

var store *sqlxstore.PostgreDatabaseStore

func TestMain(m *testing.M) {
	logger := log.New(os.Stdout, "DATABASE: ", log.Llongfile|log.LstdFlags)
	store = sqlxstore.NewPostgreDatabaseStore(logger)
	if err := store.Open(); err != nil {
		log.Fatalln(err)
	}
	defer store.Close()
	m.Run()
}
