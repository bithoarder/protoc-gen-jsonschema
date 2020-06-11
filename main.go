// protoc plugin which converts .proto to JSON schema
// It is spawned by protoc and generates JSON-schema files.
// "Heavily influenced" by Google's "protog-gen-bq-schema"
//
// usage:
//  $ bin/protoc --jsonschema_out=path/to/outdir foo.proto
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/alecthomas/jsonschema"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/xeipuuv/gojsonschema"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const (
	LOG_DEBUG = 0
	LOG_INFO  = 1
	LOG_WARN  = 2
	LOG_ERROR = 3
	LOG_FATAL = 4
	LOG_PANIC = 5
)

var (
	allowNullValues              bool = false
	disallowAdditionalProperties bool = false
	disallowBigIntsAsStrings     bool = false
	debugLogging                 bool = false
	addSchemaProperty            bool = false
	globalPkg                         = &ProtoPackage{
		name:     "",
		parent:   nil,
		children: make(map[string]*ProtoPackage),
		types:    make(map[string]*descriptor.DescriptorProto),
	}
	logLevels = map[LogLevel]string{
		0: "DEBUG",
		1: "INFO",
		2: "WARN",
		3: "ERROR",
		4: "FATAL",
		5: "PANIC",
	}
)

// ProtoPackage describes a package of Protobuf, which is an container of message types.
type ProtoPackage struct {
	name      string
	parent    *ProtoPackage
	children  map[string]*ProtoPackage
	types     map[string]*descriptor.DescriptorProto
	enumTypes map[string]*descriptor.EnumDescriptorProto
}

type LogLevel int

func init() {
	flag.BoolVar(&allowNullValues, "allow_null_values", false, "Allow NULL values to be validated")
	flag.BoolVar(&disallowAdditionalProperties, "disallow_additional_properties", false, "Disallow additional properties")
	flag.BoolVar(&disallowBigIntsAsStrings, "disallow_bigints_as_strings", false, "Disallow bigints to be strings (eg scientific notation)")
	flag.BoolVar(&debugLogging, "debug", false, "Log debug messages")
	flag.BoolVar(&addSchemaProperty, "add_schema_property", false, "Add $schema property to all messages")
}

func logWithLevel(logLevel LogLevel, logFormat string, logParams ...interface{}) {
	// If we're not doing debug logging then just return:
	if logLevel <= LOG_INFO && !debugLogging {
		return
	}

	// Otherwise log:
	logMessage := fmt.Sprintf(logFormat, logParams...)
	log.Printf(fmt.Sprintf("[%v] %v", logLevels[logLevel], logMessage))
}

func (pkg *ProtoPackage) getChildPkg(pkgName string) *ProtoPackage {
	if child, has := pkg.children[pkgName]; has {
		return child
	} else if parts := strings.SplitN(pkgName, ".", 2); len(parts) == 2 {
		return pkg.getChildPkg(parts[0]).getChildPkg(parts[1])
	} else {
		child := &ProtoPackage{
			name:      pkg.name + "." + pkgName,
			parent:    pkg,
			children:  make(map[string]*ProtoPackage),
			types:     make(map[string]*descriptor.DescriptorProto),
			enumTypes: make(map[string]*descriptor.EnumDescriptorProto),
		}

		pkg.children[pkgName] = child
		return child
	}
}

func (pkg *ProtoPackage) registerType(msg *descriptor.DescriptorProto) {
	pkg.types[msg.GetName()] = msg

	for _, enum := range msg.GetEnumType() {
		pkg.getChildPkg(msg.GetName()).enumTypes[enum.GetName()] = enum
	}

	for _, type_ := range msg.NestedType {
		pkg.getChildPkg(msg.GetName()).registerType(type_)
	}
}

func registerFile(file *descriptor.FileDescriptorProto) {
	logWithLevel(LOG_DEBUG, "Loading types package %s", file.GetPackage())

	pkg := globalPkg.getChildPkg(file.GetPackage())

	for _, msg := range file.GetMessageType() {
		logWithLevel(LOG_DEBUG, "Loading a message type %s from package %s", msg.GetName(), file.GetPackage())
		pkg.registerType(msg)
	}

	for _, enum := range file.GetEnumType() {
		logWithLevel(LOG_DEBUG, "Loading a enum type %s from package %s", enum.GetName(), file.GetPackage())
		pkg.enumTypes[enum.GetName()] = enum
	}
}

func (pkg *ProtoPackage) lookupType(name string) (*descriptor.DescriptorProto, bool) {
	if strings.HasPrefix(name, ".") {
		return globalPkg.relativelyLookupType(name[1:])
	}

	for ; pkg != nil; pkg = pkg.parent {
		if desc, ok := pkg.relativelyLookupType(name); ok {
			return desc, ok
		}
	}
	return nil, false
}

func (pkg *ProtoPackage) lookupEnumType(name string) (*descriptor.EnumDescriptorProto, bool) {
	if strings.HasPrefix(name, ".") {
		return globalPkg.relativelyLookupEnumType(name[1:])
	}

	for ; pkg != nil; pkg = pkg.parent {
		if desc, ok := pkg.relativelyLookupEnumType(name); ok {
			return desc, ok
		}
	}
	return nil, false
}

func (pkg *ProtoPackage) relativelyLookupType(name string) (*descriptor.DescriptorProto, bool) {
	components := strings.SplitN(name, ".", 2)
	switch len(components) {
	case 0:
		logWithLevel(LOG_DEBUG, "empty message name")
		return nil, false
	case 1:
		found, ok := pkg.types[components[0]]
		return found, ok
	case 2:
		logWithLevel(LOG_DEBUG, "looking for %s in %s at %s (%v)", components[1], components[0], pkg.name, pkg)
		if child, ok := pkg.children[components[0]]; ok {
			return child.relativelyLookupType(components[1])
		}
		logWithLevel(LOG_INFO, "no such package nor message %s in %s", components[0], pkg.name)
		return nil, false
	default:
		logWithLevel(LOG_FATAL, "not reached")
		return nil, false
	}
}

func (pkg *ProtoPackage) relativelyLookupEnumType(name string) (*descriptor.EnumDescriptorProto, bool) {
	components := strings.SplitN(name, ".", 2)
	switch len(components) {
	case 0:
		logWithLevel(LOG_DEBUG, "empty message name")
		return nil, false
	case 1:
		found, ok := pkg.enumTypes[components[0]]
		return found, ok
	case 2:
		logWithLevel(LOG_DEBUG, "looking for %s in %s at %s (%v)", components[1], components[0], pkg.name, pkg)
		if child, ok := pkg.children[components[0]]; ok {
			return child.relativelyLookupEnumType(components[1])
		}
		logWithLevel(LOG_INFO, "no such package nor message %s in %s", components[0], pkg.name)
		return nil, false
	default:
		logWithLevel(LOG_FATAL, "not reached")
		return nil, false
	}
}

func (pkg *ProtoPackage) relativelyLookupPackage(name string) (*ProtoPackage, bool) {
	components := strings.Split(name, ".")
	for _, c := range components {
		var ok bool
		pkg, ok = pkg.children[c]
		if !ok {
			return nil, false
		}
	}
	return pkg, true
}

func setMinimum(jsonSchemaType *jsonschema.Type, min int) {
	if min != 0 {
		jsonSchemaType.Minimum = int(min)
	} else {
		// a hack to get around omitempty on Minimum
		jsonSchemaType.Minimum = -1
		jsonSchemaType.ExclusiveMinimum = true
	}
}

func setMaximum(jsonSchemaType *jsonschema.Type, max int) {
	if max != 0 {
		jsonSchemaType.Maximum = int(max)
	} else {
		// a hack to get around omitempty on Maximum
		jsonSchemaType.Maximum = 1
		jsonSchemaType.ExclusiveMaximum = true
	}
}

// Convert a proto "field" (essentially a type-switch with some recursion):
func convertField(curPkg *ProtoPackage, desc *descriptor.FieldDescriptorProto, msg *descriptor.DescriptorProto) (*jsonschema.Type, bool, error) {

	// Prepare a new jsonschema.Type for our eventual return value:
	jsonSchemaType := &jsonschema.Type{
		Properties: make(map[string]*jsonschema.Type),
	}

	required := false

	// Switch the types, and pick a JSONSchema equivalent:
	switch desc.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_NUMBER},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_NUMBER
		}

	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SINT32:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_INTEGER},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_INTEGER
		}

	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT64:
		jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_INTEGER})
		if !disallowBigIntsAsStrings {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		}
		if allowNullValues {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_NULL})
		}

	case descriptor.FieldDescriptorProto_TYPE_STRING,
		descriptor.FieldDescriptorProto_TYPE_BYTES:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_STRING},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
		}

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_STRING})
		jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_INTEGER})
		if allowNullValues {
			jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_NULL})
		}

		if enumType, has := curPkg.lookupEnumType(desc.GetTypeName()); has {
			for _, enumValue := range enumType.Value {
				jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Name)
				jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Number)
			}
		} else {
			return nil, false, fmt.Errorf("no such enum type named %s", desc.GetTypeName())
		}

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_BOOLEAN},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_BOOLEAN
		}

	case descriptor.FieldDescriptorProto_TYPE_GROUP,
		descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
		//if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL {
		//	jsonSchemaType.AdditionalProperties = []byte("true")
		//}
		//if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED {
		//	jsonSchemaType.AdditionalProperties = []byte("false")
		//}
		jsonSchemaType.AdditionalProperties = []byte("false")

	default:
		return nil, false, fmt.Errorf("unrecognized field type: %s", desc.GetType().String())
	}

	isRepeated := desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED

	if desc.Options != nil {
		if ext, err := proto.GetExtension(desc.Options, E_JsonSchema); err == proto.ErrMissingExtension {
		} else if err != nil {
			return nil, false, err
		} else if jsonSchema, ok := ext.(*Schema); ok {
			jsonSchemaType.Description = jsonSchema.Description
			jsonSchemaType.MinLength = int(jsonSchema.MinLength)
			jsonSchemaType.MaxLength = int(jsonSchema.MaxLength)
			jsonSchemaType.MinItems = int(jsonSchema.MinItems)
			jsonSchemaType.MaxItems = int(jsonSchema.MaxItems)
			jsonSchemaType.UniqueItems = jsonSchema.UniqueItems

			required = jsonSchema.Required

			switch r := jsonSchema.Rule.(type) {
			case nil:
				// only had an description
			case *Schema_Ref:
				// replace whole schema with just a ref
				return &jsonschema.Type{Ref: r.Ref}, required, nil
			case *Schema_Pattern:
				jsonSchemaType.Pattern = r.Pattern
			case *Schema_Min:
				setMinimum(jsonSchemaType, int(r.Min))
			case *Schema_Max:
				setMaximum(jsonSchemaType, int(r.Max))
			case *Schema_Range_:
				setMinimum(jsonSchemaType, int(r.Range.Min))
				setMaximum(jsonSchemaType, int(r.Range.Max))
			default:
				return nil, false, fmt.Errorf("unsupported json schema extension %T", r)
			}
		}
	}

	// Recurse array of primitive types:
	if isRepeated && jsonSchemaType.Type != gojsonschema.TYPE_OBJECT {
		jsonSchemaType.Items = &jsonschema.Type{}

		if len(jsonSchemaType.Enum) > 0 {
			jsonSchemaType.Items.Enum = jsonSchemaType.Enum
			jsonSchemaType.Enum = nil
			jsonSchemaType.Items.OneOf = nil
		} else {
			jsonSchemaType.Items.Type = jsonSchemaType.Type
			jsonSchemaType.Items.OneOf = jsonSchemaType.OneOf
		}

		jsonSchemaType.Items.Pattern, jsonSchemaType.Pattern = jsonSchemaType.Pattern, jsonSchemaType.Items.Pattern
		jsonSchemaType.Items.Minimum, jsonSchemaType.Minimum = jsonSchemaType.Minimum, jsonSchemaType.Items.Minimum
		jsonSchemaType.Items.Maximum, jsonSchemaType.Maximum = jsonSchemaType.Maximum, jsonSchemaType.Items.Maximum
		jsonSchemaType.Items.ExclusiveMinimum, jsonSchemaType.ExclusiveMinimum = jsonSchemaType.ExclusiveMinimum, jsonSchemaType.Items.ExclusiveMinimum
		jsonSchemaType.Items.ExclusiveMaximum, jsonSchemaType.ExclusiveMaximum = jsonSchemaType.ExclusiveMaximum, jsonSchemaType.Items.ExclusiveMaximum

		if allowNullValues {
			jsonSchemaType.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NULL},
				{Type: gojsonschema.TYPE_ARRAY},
			}
		} else {
			jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
			jsonSchemaType.OneOf = []*jsonschema.Type{}
		}

		return jsonSchemaType, required, nil
	}

	// Recurse nested objects / arrays of objects (if necessary):
	if jsonSchemaType.Type == gojsonschema.TYPE_OBJECT {
		//jsonSchemaType.Description = fmt.Sprintf("%s", desc.GetTypeName())

		// todo: find way to configure known types, maybe via protobuf message extensions, for now this is just a quick hack:
		if desc.GetTypeName() == ".google.protobuf.Timestamp" {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
			jsonSchemaType.AdditionalProperties = nil
			// this is the only allowed format, see https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/timestamp.proto
			jsonSchemaType.Pattern = `^20[0-9]{2}-([0]?[1-9]|[1][0-2])-([0]?[1-9]|[12][0-9]|3[01])T([0]?[0-9]|1[0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?Z$`
		} else if desc.GetTypeName() == ".google.protobuf.Duration" {
			jsonSchemaType.Type = gojsonschema.TYPE_STRING
			jsonSchemaType.AdditionalProperties = nil
			// simplified version of Duration parsing:
			jsonSchemaType.Pattern = `^[-+]?(([0-9]+h)([0-5]?[0-9]m)?([0-5]?[0-9]s)?([0-9]?[0-9]?[0-9]ms)?|([0-9]+m)([0-5]?[0-9]s)?([0-9]?[0-9]?[0-9]ms)?|([0-9]+s)([0-9]?[0-9]?[0-9]ms)?|([0-9]+ms))$`
		} else {
			recordType, ok := curPkg.lookupType(desc.GetTypeName())
			if !ok {
				return nil, false, fmt.Errorf("no such message type named %s", desc.GetTypeName())
			}

			// Recurse:
			recursedJSONSchemaType, err := convertMessageType(curPkg, recordType)
			if err != nil {
				return nil, false, err
			}

			// The result is stored differently for arrays of objects (they become "items"):
			if desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
				jsonSchemaType.Items = &recursedJSONSchemaType
				jsonSchemaType.Type = gojsonschema.TYPE_ARRAY
			} else {
				// Nested objects are more straight-forward:
				jsonSchemaType.Properties = recursedJSONSchemaType.Properties
				jsonSchemaType.Required = recursedJSONSchemaType.Required
			}

			// Optionally allow NULL values:
			if allowNullValues {
				jsonSchemaType.OneOf = []*jsonschema.Type{
					{Type: gojsonschema.TYPE_NULL},
					{Type: jsonSchemaType.Type},
				}
				jsonSchemaType.Type = ""
			}
		}
	}

	return jsonSchemaType, required, nil
}

// Converts a proto "MESSAGE" into a JSON-Schema:
func convertMessageType(curPkg *ProtoPackage, msg *descriptor.DescriptorProto) (jsonschema.Type, error) {

	// Prepare a new jsonschema:
	jsonSchemaType := jsonschema.Type{
		Properties: make(map[string]*jsonschema.Type),
		Version:    jsonschema.Version,
	}

	// Optionally allow NULL values:
	if allowNullValues {
		jsonSchemaType.OneOf = []*jsonschema.Type{
			{Type: gojsonschema.TYPE_NULL},
			{Type: gojsonschema.TYPE_OBJECT},
		}
	} else {
		jsonSchemaType.Type = gojsonschema.TYPE_OBJECT
	}

	// disallowAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	if disallowAdditionalProperties {
		jsonSchemaType.AdditionalProperties = []byte("false")
	} else {
		jsonSchemaType.AdditionalProperties = []byte("true")
	}

	if addSchemaProperty {
		jsonSchemaType.Properties["$schema"] = &jsonschema.Type{
			Type:   "string",
			Format: "uri",
		}
	}

	oneOf := map[int32]*jsonschema.Type{}

	logWithLevel(LOG_DEBUG, "Converting message: %s", proto.MarshalTextString(msg))
	for _, fieldDesc := range msg.GetField() {
		if fieldDesc.Options.GetDeprecated() {
			logWithLevel(LOG_DEBUG, "Field %s in %s is deprecated", fieldDesc.GetName(), msg.GetName())
		} else {
			if jsonSchemaType.AnyOf == nil {
				jsonSchemaType.AnyOf = []*jsonschema.Type{}
				jsonSchemaType.AnyOf = append(jsonSchemaType.AnyOf, jsonSchemaType.OneOf...)
				jsonSchemaType.OneOf = nil
			}

			if fieldDesc.OneofIndex != nil {
				if oneOfType, has := oneOf[*fieldDesc.OneofIndex]; !has {
					oneOfType = &jsonschema.Type{
						//Properties: make(map[string]*jsonschema.Type),
						OneOf: []*jsonschema.Type{},
					}
					jsonSchemaType.AnyOf = append(jsonSchemaType.AnyOf, oneOfType)
					oneOf[*fieldDesc.OneofIndex] = oneOfType

					oneOfType.OneOf = append(oneOfType.OneOf, &jsonschema.Type{Required: []string{*fieldDesc.JsonName}})
				} else {
					oneOfType.OneOf = append(oneOfType.OneOf, &jsonschema.Type{Required: []string{*fieldDesc.JsonName}})
				}
			}

			recursedJSONSchemaType, required, err := convertField(curPkg, fieldDesc, msg)
			if err != nil {
				logWithLevel(LOG_ERROR, "Failed to convert field %s in %s: %v", fieldDesc.GetName(), msg.GetName(), err)
				return jsonSchemaType, err
			}
			jsonSchemaType.Properties[*fieldDesc.JsonName] = recursedJSONSchemaType
			if required {
				jsonSchemaType.Required = append(jsonSchemaType.Required, *fieldDesc.JsonName)
			}
		}
	}
	return jsonSchemaType, nil
}

// Converts a proto "ENUM" into a JSON-Schema:
func convertEnumType(enum *descriptor.EnumDescriptorProto) (jsonschema.Type, error) {

	// Prepare a new jsonschema.Type for our eventual return value:
	jsonSchemaType := jsonschema.Type{
		Version: jsonschema.Version,
	}

	// Allow both strings and integers:
	jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: "string"})
	jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: "integer"})

	// Add the allowed values:
	for _, enumValue := range enum.Value {
		jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Name)
		jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Number)
	}

	return jsonSchemaType, nil
}

// Converts a proto file into a JSON-Schema:
func convertFile(file *descriptor.FileDescriptorProto) ([]*plugin.CodeGeneratorResponse_File, error) {

	// Input filename:
	protoFileName := path.Base(file.GetName())

	// Prepare a list of responses:
	response := []*plugin.CodeGeneratorResponse_File{}

	// Warn about multiple messages / enums in files:
	if len(file.GetMessageType()) > 1 {
		logWithLevel(LOG_INFO, "protoc-gen-jsonschema will create multiple MESSAGE schemas (%d) from one proto file (%v)", len(file.GetMessageType()), protoFileName)
	}
	if len(file.GetEnumType()) > 1 {
		logWithLevel(LOG_INFO, "protoc-gen-jsonschema will create multiple ENUM schemas (%d) from one proto file (%v)", len(file.GetEnumType()), protoFileName)
	}

	// Generate standalone ENUMs:
	if len(file.GetMessageType()) == 0 {
		for _, enum := range file.GetEnumType() {
			jsonSchemaFileName := fmt.Sprintf("%s.jsonschema", enum.GetName())
			logWithLevel(LOG_INFO, "Generating JSON-schema for stand-alone ENUM (%v) in file [%v] => %v", enum.GetName(), protoFileName, jsonSchemaFileName)
			enumJsonSchema, err := convertEnumType(enum)
			if err != nil {
				logWithLevel(LOG_ERROR, "Failed to convert %s: %v", protoFileName, err)
				return nil, err
			} else {
				// Marshal the JSON-Schema into JSON:
				jsonSchemaJSON, err := json.MarshalIndent(enumJsonSchema, "", "    ")
				if err != nil {
					logWithLevel(LOG_ERROR, "Failed to encode jsonSchema: %v", err)
					return nil, err
				} else {
					// Add a response:
					resFile := &plugin.CodeGeneratorResponse_File{
						Name:    proto.String(jsonSchemaFileName),
						Content: proto.String(string(jsonSchemaJSON)),
					}
					response = append(response, resFile)
				}
			}
		}
	} else {
		// Otherwise process MESSAGES (packages):
		pkg, ok := globalPkg.relativelyLookupPackage(file.GetPackage())
		if !ok {
			return nil, fmt.Errorf("no such package found: %s", file.GetPackage())
		}
		for _, msg := range file.GetMessageType() {
			jsonSchemaFileName := fmt.Sprintf("%s.schema.json", msg.GetName())
			logWithLevel(LOG_INFO, "Generating JSON-schema for MESSAGE (%v) in file [%v] => %v", msg.GetName(), protoFileName, jsonSchemaFileName)
			messageJSONSchema, err := convertMessageType(pkg, msg)
			if err != nil {
				logWithLevel(LOG_ERROR, "Failed to convert %s: %v", protoFileName, err)
				return nil, err
			} else {
				// Marshal the JSON-Schema into JSON:
				jsonSchemaJSON, err := json.MarshalIndent(messageJSONSchema, "", "    ")
				if err != nil {
					logWithLevel(LOG_ERROR, "Failed to encode jsonSchema: %v", err)
					return nil, err
				} else {
					// Add a response:
					resFile := &plugin.CodeGeneratorResponse_File{
						Name:    proto.String(jsonSchemaFileName),
						Content: proto.String(string(jsonSchemaJSON)),
					}
					response = append(response, resFile)
				}
			}
		}
	}

	return response, nil
}

func convert(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	generateTargets := make(map[string]bool)
	for _, file := range req.GetFileToGenerate() {
		generateTargets[file] = true
	}

	res := &plugin.CodeGeneratorResponse{}
	for _, file := range req.GetProtoFile() {
		registerFile(file)
	}

	for _, file := range req.GetProtoFile() {
		if _, ok := generateTargets[file.GetName()]; ok {
			logWithLevel(LOG_DEBUG, "Converting file (%v)", file.GetName())
			converted, err := convertFile(file)
			if err != nil {
				panic(err)
				res.Error = proto.String(fmt.Sprintf("Failed to convert %s: %v", file.GetName(), err))
				return res, err
			}
			res.File = append(res.File, converted...)
		}
	}
	return res, nil
}

func convertFrom(rd io.Reader) (*plugin.CodeGeneratorResponse, error) {
	logWithLevel(LOG_DEBUG, "Reading code generation request")
	input, err := ioutil.ReadAll(rd)
	if err != nil {
		logWithLevel(LOG_ERROR, "Failed to read request: %v", err)
		return nil, err
	}

	req := &plugin.CodeGeneratorRequest{}
	err = proto.Unmarshal(input, req)
	if err != nil {
		logWithLevel(LOG_ERROR, "Can't unmarshal input: %v", err)
		return nil, err
	}

	commandLineParameter(req.GetParameter())

	logWithLevel(LOG_DEBUG, "Converting input")
	return convert(req)
}

func commandLineParameter(parameters string) {
	for _, parameter := range strings.Split(parameters, ",") {
		switch parameter {
		case "allow_null_values":
			allowNullValues = true
		case "debug":
			debugLogging = true
		case "disallow_additional_properties":
			disallowAdditionalProperties = true
		case "disallow_bigints_as_strings":
			disallowBigIntsAsStrings = true
		case "add_schema_property":
			addSchemaProperty = true
		}
	}
}

func main() {
	flag.Parse()
	ok := true
	logWithLevel(LOG_DEBUG, "Processing code generator request")
	res, err := convertFrom(os.Stdin)
	if err != nil {
		ok = false
		if res == nil {
			message := fmt.Sprintf("Failed to read input: %v", err)
			res = &plugin.CodeGeneratorResponse{
				Error: &message,
			}
		}
	}

	logWithLevel(LOG_DEBUG, "Serializing code generator response")
	data, err := proto.Marshal(res)
	if err != nil {
		logWithLevel(LOG_FATAL, "Cannot marshal response: %v", err)
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		logWithLevel(LOG_FATAL, "Failed to write response: %v", err)
	}

	if ok {
		logWithLevel(LOG_DEBUG, "Succeeded to process code generator request")
	} else {
		logWithLevel(LOG_WARN, "Failed to process code generator but successfully sent the error to protoc")
		os.Exit(1)
	}
}
