package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func parseReq(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var req plugin.CodeGeneratorRequest
	if err = proto.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func mapKeys(m map[string]string) []string {
	ks := []string{}
	for k, _ := range m {
		ks = append(ks, k)
	}
	return ks
}

func processReq(req *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	var resp plugin.CodeGeneratorResponse
	for _, file := range req.ProtoFile {
		for _, messageType := range file.GetMessageType() {
			if len(messageType.GetNestedType()) > 0 {
				panic("not support nested type")
			}
			buf := new(bytes.Buffer)
			var oneofTypes []map[string]string
			for range messageType.GetOneofDecl() {
				oneofTypes = append(oneofTypes, make(map[string]string))
			}
			fmt.Fprintf(buf, "type %s = Readonly<{\n", messageType.GetName())
			for _, field := range messageType.GetField() {
				key := field.GetJsonName()
				tsType := convertType(field)
				if field.OneofIndex != nil {
					oneofTypes[field.GetOneofIndex()][key] = tsType
				} else {
					fmt.Fprintf(buf, "  %s: %s;\n", key, tsType)
				}
			}
			fmt.Fprintf(buf, "}>")
			writeOneofTypes(buf, oneofTypes)
			fmt.Fprintf(buf, ";\n")
			appendBufferToFile(&resp, messageType.GetName(), buf)
		}
		writeEnumType(&resp, *file)
	}
	return &resp
}

func appendBufferToFile(resp *plugin.CodeGeneratorResponse, name string, buf *bytes.Buffer) {
	resp.File = append(resp.File, &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(name + ".d.ts"),
		Content: proto.String(buf.String()),
	})
}

func writeEnumType(resp *plugin.CodeGeneratorResponse, file descriptor.FileDescriptorProto) {
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
		appendBufferToFile(resp, enumType.GetName(), buf)
	}
}

func writeOneofTypes(buf *bytes.Buffer, oneofTypes []map[string]string) {
	for index, oneof := range oneofTypes {
		if index == 0 {
			fmt.Fprintf(buf, " &\n")
		}
		fmt.Fprintf(buf, "  Readonly<\n")
		oneofKeys := mapKeys(oneof)
		for oneofIndex, oneofKey := range oneofKeys {
			fmt.Fprintf(buf, "    {\n")
			for key, tsType := range oneof {
				if oneofKey == key {
					fmt.Fprintf(buf, "      %s?: %s;\n", key, tsType)
				} else {
					fmt.Fprintf(buf, "      %s?: never;\n", key)
				}
			}
			fmt.Fprintf(buf, "    }")
			if len(oneofKeys)-1 == oneofIndex {
				fmt.Fprintf(buf, "\n")
			} else {
				fmt.Fprintf(buf, " |\n")
			}
		}
		if len(oneofTypes)-1 == index {
			fmt.Fprintf(buf, "  >")
		} else {
			fmt.Fprintf(buf, "  > &\n")
		}
	}
}

func convertType(field *descriptor.FieldDescriptorProto) string {
	var tsType string
	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE,
		descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		tsType = "number"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		tsType = "boolean"
	case descriptor.FieldDescriptorProto_TYPE_STRING,
		descriptor.FieldDescriptorProto_TYPE_BYTES:
		tsType = "string"
	case descriptor.FieldDescriptorProto_TYPE_ENUM,
		descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		ns := strings.Split(field.GetTypeName(), ".")
		tsType = ns[len(ns)-1]
	default:
		tsType = "unknown"
	}
	if field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		tsType = "ReadonlyArray<" + tsType + ">"
	}
	return tsType
}

func emitResp(resp *plugin.CodeGeneratorResponse) error {
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

	resp := processReq(req)

	return emitResp(resp)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
