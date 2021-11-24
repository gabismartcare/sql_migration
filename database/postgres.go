package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Postgres connection
	"github.com/sql-migration/model"
	environment "github.com/sql-migration/tool"
	"log"
	"strconv"
	"strings"
	"time"
)

type PgConnection struct {
	*sql.DB
	waiting string
	files   []string
}

type Postgres struct {
	URL      string
	Username string
	Password string
	Port     string
	Database string
}

func testConnection(p *Postgres) error {
	psqlInfo := getPsqlInfo(p)
	conn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Exec("SELECT 1 ")
	if err != nil {
		return err
	}
	return nil
}
func (p *Postgres) WaitFor(timeoutInSeconds int64) bool {
	tick := time.Tick(1 * time.Second)
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Duration(timeoutInSeconds) * time.Second)
		timeout <- true
	}()
	for {
		select {
		case _ = <-tick:
			log.Printf("Attempting to connect to database %v", p)
			if err := testConnection(p); err == nil {
				return true
			} else {
				log.Println(err)
			}
		case _ = <-timeout:
			return false
		}
	}
}
func DefaultConf() *Postgres {
	return &Postgres{
		Username: environment.GetOr("POSTGRES_USERNAME", "postgres"),
		Password: environment.GetOr("POSTGRES_PASSWORD", "postgres"),
		URL:      environment.GetOr("POSTGRES_URL", "127.0.0.1"),
		Database: environment.GetOr("POSTGRES_DATABASE", ""),
		Port:     environment.GetOr("POSTGRES_PORT", "5432"),
	}
}
func getPsqlInfo(conf *Postgres) string {

	url := conf.URL
	port := 5432
	if strings.HasPrefix(conf.URL, "tcp") {
		url = strings.ReplaceAll(conf.URL, "tcp://", "")
	}

	splittedUrl := strings.Split(url, ":")

	if len(splittedUrl) == 1 {
		url = splittedUrl[0]
		var err error
		port, err = strconv.Atoi(conf.Port)
		if err != nil {
			log.Print("Cannot parse " + conf.Port + " using 5432")
			port = 5432
		}
	} else if len(splittedUrl) == 2 {
		url = splittedUrl[0]
		var err error
		port, err = strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Print("Cannot parse " + splittedUrl[1] + "trying config")
		}
		if len(conf.Port) != 0 {
			port, err = strconv.Atoi(conf.Port)
			if err != nil {
				log.Print("Cannot parse " + conf.Port + "trying config")
			}
		}
	}
	if port == 0 {
		log.Println("port is not set, using 5432")
		port = 5432
	}
	psqlInfo := ""
	if len(conf.Password) != 0 {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable application_name='golang'", url, port, conf.Username, conf.Password, conf.Database)
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable application_name='golang'", url, port, conf.Username, conf.Database)
	}
	return psqlInfo
}

func GetPgConnection(conf *Postgres) (PgConnection, error) {
	log.Println("starting postgres driver")
	if conf.WaitFor(30) {
		psqlInfo := getPsqlInfo(conf)
		conn, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			return PgConnection{}, err
		}
		log.Println("started")
		return PgConnection{DB: conn}, nil
	}
	return PgConnection{}, fmt.Errorf("timeout when connecting to %v", conf)
}

func (p *PgConnection) CreateChangelogTableIfNotExists() error {
	if _, err := p.DB.Exec("CREATE TABLE IF NOT EXISTS changelog(fileName varchar(256), created timestamp)"); err != nil {
		return err
	}
	return nil
}

func (p *PgConnection) GetAllFilesAlreadyImported() ([]model.File, error) {
	r, err := p.DB.Query("SELECT fileName from changelog order by created asc")
	if err != nil {
		return nil, err
	}
	defer r.Close()

	files := make([]model.File, 0, 10)
	for r.Next() {
		var fileName, md5 string
		if err := r.Scan(&fileName, &md5); err != nil {
			return nil, err
		}
		files = append(files, model.File{
			Name: fileName,
			Md5:  md5,
		})
	}
	return files, nil
}

func (p *PgConnection) AddForTransaction(fileName string, s string) {
	p.waiting += s
}

func (p *PgConnection) ApplyChanges() error {
	tx, err := p.DB.Begin()
	if err != nil {
		return err
	}
	rollback := false
	defer func() {
		if rollback {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	if _, err := tx.Exec(p.waiting); err != nil {
		rollback = true
		return err
	}
	for _, f := range p.files {
		f := f
		if _, err := tx.Exec("INSERT into changelog (file) VALUES (?)", f); err != nil {
			rollback = true
			return err
		}
	}
	return err
}
