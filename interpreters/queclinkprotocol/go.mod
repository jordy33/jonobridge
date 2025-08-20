module queclinkprotocol

go 1.23.2

require (
	github.com/MaddSystems/jonobridge/common v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Replace directive pointing to the local common module
replace github.com/MaddSystems/jonobridge/common => ../../common
