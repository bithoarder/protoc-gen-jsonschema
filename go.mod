module github.com/chrusty/protoc-gen-jsonschema

require (
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921
	github.com/golang/protobuf v1.3.2
	github.com/iancoleman/orderedmap v0.0.0-20190318233801-ac98e3ecb4b0
	github.com/sirupsen/logrus v1.1.0
	github.com/stretchr/testify v1.3.1-0.20190311161405-34c6fa2dc709
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20181006164115-f58b4a9e3d67
)

//replace github.com/alecthomas/jsonschema => github.com/bithoarder/jsonschema v0.0.0-20200530073317-71f438968921
//replace github.com/alecthomas/jsonschema => c:\proj\go\src\github.com\bithoarder\jsonschema

go 1.13
