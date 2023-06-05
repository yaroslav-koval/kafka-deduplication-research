# Test utils package

This is a package that helps and simplifies the job of writing tests.

At the moment, this package helps to run containers in the docker. To run PostgreSQL, Elasticsearch and MongoDB there are structures to make things easier.

## Test running command

Current test running command:
```
go test -coverpkg=./... -coverprofile=profile.cov.all -p=1 ./...
```
This command run test with generation coverage profile for calculate coverage percent and with paralell run with count = 5
