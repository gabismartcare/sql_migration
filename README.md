# Sql Migration Tool

Simplistic docker migration database tool

## Allowed environment variables

````
POSTGRES_USERNAME postgres
POSTGRES_PASSWORD password
POSTGRES_URL localhost
POSTGRES_DATABASE 
POSTGRES_PORT 5432
INPUT_DIR /migration
````

## Changelog

In the ${INPUT_DIR} directory put a changelog.yml

````
changelog:
  - change:
      file: 01-create-user.sql
  - change:
      file: 02-update-user-table.sql
````

## Running

### From commandline

````
docker run -rm -v ${PWD}:/migration gabismartcare/sql_migration
````

### Build new docker image

````
FROM gabismartcare/sqlmigration
COPY *.sql /migration