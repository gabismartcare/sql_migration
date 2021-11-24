package model

import (
	"errors"
	"github.com/sql-migration/storage"
	"testing"
)

func TestChangelogFromStorage(t *testing.T) {
	c, err := ChangelogFromStorage(storage.FileStorage{Directory: "./"})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Changes) != 2 {
		t.Fatal(errors.New("Expecting 2 changes"))
	}
}
