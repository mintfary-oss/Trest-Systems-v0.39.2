module github.com/mintfary-oss/trest-sistems

go 1.23.0

require (
	github.com/jackc/pgx/v5 v5.7.5
	github.com/spf13/cobra v1.10.2
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/spf13/cobra => ./third_party/cobra

replace gopkg.in/yaml.v3 => ./third_party/yamlv3

replace github.com/jackc/pgx/v5 => ./third_party/pgx
