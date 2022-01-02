package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func parseReq(r io.Reader) (*pluginpb.CodeGeneratorRequest, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var req pluginpb.CodeGeneratorRequest
	if err = proto.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func processReq(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	var err error
	var resp pluginpb.CodeGeneratorResponse
	for _, file := range req.ProtoFile {
		for _, messageType := range file.GetMessageType() {
			buf := new(bytes.Buffer)
			fields := make(map[string]string)
			fieldsOrder := []string{}
			nestedTypes := make(map[string]string)
			messageTypeName := messageType.GetName()

			switch messageTypeName {
			case "DoubleValue",
				"FloatValue",
				"Int32Value",
				"Int64Value",
				"UInt32Value",
				"UInt64Value",
				"BoolValue",
				"BytesValue",
				"StringValue":
				continue
			}

			oneofTypes := []map[string]string{}
			oneofTypesOrder := [][]string{}
			for range messageType.GetOneofDecl() {
				oneofTypes = append(oneofTypes, make(map[string]string))
				oneofTypesOrder = append(oneofTypesOrder, make([]string, 0))
			}
			for _, nestedType := range messageType.GetNestedType() {
				options := nestedType.GetOptions()
				if options != nil && options.GetMapEntry() {
					name := nestedType.GetName()
					key := convertType(nestedType.GetField()[0], nestedTypes)
					value := convertType(nestedType.GetField()[1], nestedTypes)
					nestedTypes[name] = fmt.Sprintf("Record<%s, %s>", key, value)
				} else {
					err = errors.New("not support nested type")
				}
			}
			for _, field := range messageType.GetField() {
				key := field.GetJsonName()
				tsType := convertType(field, nestedTypes)
				if field.OneofIndex != nil {
					oneofIndex := field.GetOneofIndex()
					oneofTypes[oneofIndex][key] = tsType
					oneofTypesOrder[oneofIndex] = append(oneofTypesOrder[oneofIndex], key)
				} else {
					fields[key] = tsType
					fieldsOrder = append(fieldsOrder, key)
				}
			}
			if len(fieldsOrder) > 0 {
				fmt.Fprintf(buf, "type %s = Readonly<{\n", messageTypeName)
				for _, key := range fieldsOrder {
					fmt.Fprintf(buf, "  %s: %s;\n", key, fields[key])
				}
				fmt.Fprintf(buf, "}>")
				if len(oneofTypes) > 0 {
					fmt.Fprintf(buf, " &\n")
				}
			} else if len(oneofTypes) > 0 {
				fmt.Fprintf(buf, "type %s =\n", messageTypeName)
			} else {
				fmt.Fprintf(buf, "type %s = null", messageTypeName)
			}
			for index, oneof := range oneofTypes {
				fmt.Fprintf(buf, "  Readonly<\n")
				for _, oneofKey := range oneofTypesOrder[index] {
					fmt.Fprintf(buf, "    | {\n")
					for _, key := range oneofTypesOrder[index] {
						if oneofKey == key {
							fmt.Fprintf(buf, "        %s?: %s;\n", key, oneof[key])
						} else {
							fmt.Fprintf(buf, "        %s?: never;\n", key)
						}
					}
					fmt.Fprintf(buf, "      }\n")
				}
				if len(oneofTypes)-1 == index {
					fmt.Fprintf(buf, "  >")
				} else {
					fmt.Fprintf(buf, "  > &\n")
				}
			}
			fmt.Fprintf(buf, ";\n")
			appendBufferToFile(&resp, messageTypeName, buf)
		}
		for _, enumType := range file.GetEnumType() {
			buf := new(bytes.Buffer)
			fmt.Fprintf(buf, "type %s =\n", enumType.GetName())
			valueLength := len(enumType.GetValue())
			for index, value := range enumType.GetValue() {
				fmt.Fprintf(buf, "  | '%s'", value.GetName())
				if index == valueLength-1 {
					fmt.Fprintf(buf, ";")
				}
				fmt.Fprintf(buf, "\n")
			}
			appendBufferToFile(&resp, enumType.GetName(), buf)
		}
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func appendBufferToFile(resp *pluginpb.CodeGeneratorResponse, name string, buf *bytes.Buffer) {
	resp.File = append(resp.File, &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(name + ".d.ts"),
		Content: proto.String(buf.String()),
	})
}

func convertType(field *descriptorpb.FieldDescriptorProto, nestedTypes map[string]string) string {
	var tsType string
	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		tsType = "number"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		tsType = "boolean"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING,
		descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		tsType = "string"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		ns := strings.Split(field.GetTypeName(), ".")
		tsType = ns[len(ns)-1]

		switch tsType {
		case "DoubleValue", "FloatValue", "Int32Value", "Int64Value", "UInt32Value", "UInt64Value":
			tsType = "number | undefined"
		case "BoolValue":
			tsType = "boolean | undefined"
		case "BytesValue", "StringValue":
			tsType = "string | undefined"
		}
	default:
		tsType = "unknown"
	}
	if val, ok := nestedTypes[tsType]; ok {
		tsType = "Readonly<" + val + ">"
	} else if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		tsType = "ReadonlyArray<" + tsType + ">"
	}
	return tsType
}

func emitResp(resp *pluginpb.CodeGeneratorResponse) error {
	buf, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(buf)
	return err
}

func run() error {
	req, err := parseReq(os.Stdin)
	if err != nil {
		return err
	}

	resp, err := processReq(req)

	if err != nil {
		return err
	}

	return emitResp(resp)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
