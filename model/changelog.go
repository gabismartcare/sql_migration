package model

import (
	"github.com/sql-migration/storage"
	"gopkg.in/yaml.v3"
)

type ChangeLogContainer struct {
	Changes []Changelog `yaml:"changelog"`
}
type Changelog struct {
	Changes Change `yaml:"change"`
}
type Change struct {
	File string `yaml:"file"`
}

func ChangelogFromStorage(storage storage.FileStorage) (ChangeLogContainer, error) {
	data, err := storage.Read("changelog.yml")
	if err != nil {
		return ChangeLogContainer{}, err
	}
	c := ChangeLogContainer{}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return ChangeLogContainer{}, err
	}
	return c, nil
}
