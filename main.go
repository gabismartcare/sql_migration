package main

import (
	"fmt"
	"github.com/sql-migration/database"
	"github.com/sql-migration/model"
	"github.com/sql-migration/storage"
	environment "github.com/sql-migration/tool"
	"log"
	"net/http"
)

func main() {
	pgConnection, err := database.GetPgConnection(database.DefaultConf())
	if err != nil {
		log.Fatal(err)
	}
	fileStorage := storage.FileStorage{Directory: environment.GetOr("INPUT_DIR", "/migration")}

	if err := pgConnection.CreateChangelogTableIfNotExists(); err != nil {
		log.Fatal(err)
	}
	files, err := pgConnection.GetAllFilesAlreadyImported()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d files already imported", len(files))
	c, err := model.ChangelogFromStorage(fileStorage)
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	var f model.File
	for i, f = range files {
		if c.Changes[i].Changes.File != f.Name {
			log.Fatal(fmt.Println("file order has change"))
		}
	}
	for ; i < len(c.Changes); i++ {
		log.Printf("loading new file %s", c.Changes[i].Changes.File)
		data, err := fileStorage.Read(c.Changes[i].Changes.File)
		if err != nil {
			log.Fatal(err)
		}
		pgConnection.AddForTransaction(c.Changes[i].Changes.File, string(data))
	}
	if err := pgConnection.ApplyChanges(); err != nil {
		log.Fatal(err)
	}
	log.Println("done")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", environment.GetOr("PORT", "8080")), nil))
}
