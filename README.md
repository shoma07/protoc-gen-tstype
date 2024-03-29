# protoc-gen-tstype

Generate TypeScript Readonly type from Proto file.

## Install

```
$ go install github.com/shoma07/protoc-gen-tstype
```

## Run

```
$ protoc -I. --plugin=protoc-gen-tstype --tstype_out=. hoge.proto
```

## Types

### Message

```proto
message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
}
```

```typescript
type SearchRequest = Readonly<{
  query: string;
  pageNumber: number;
  resultPerPage: number;
}>;
```

### Repeated

```proto
message SearchResponse {
  repeated Result results = 1;
}
```

```typescript
type SearchResponse = Readonly<{
  results: ReadonlyArray<Result>;
}>;
```

### Enum

```proto
enum Corpus {
  UNIVERSAL = 0;
  WEB = 1;
  IMAGES = 2;
  LOCAL = 3;
  NEWS = 4;
  PRODUCTS = 5;
  VIDEO = 6;
}
```

```typescript
type Corpus =
  | 'UNIVERSAL'
  | 'WEB'
  | 'IMAGES'
  | 'LOCAL'
  | 'NEWS'
  | 'PRODUCTS'
  | 'VIDEO';
```

### Oneof

```proto
message SampleMessage {
  oneof test_oneof {
    string name = 4;
    SubMessage sub_message = 9;
  }
}
```

```typescript
type SampleMessage =
  Readonly<
    {
      name?: string;
      subMessage?: never;
    } |
    {
      name?: never;
      subMessage?: SubMessage;
    }
  >;
```

### Maps

```proto
message SampleMessage {
  map<string, Project> projects = 1;
}
```

```typescript
type SampleMessage = Readonly<{
  projects: Readonly<Record<string, Project>>;
}>;
```

### Nested Types

NestedType other than map is not supported.
Define message without nesting it.

### Wrappers

```proto
message SampleMessage {
  google.protobuf.DoubleValue double_value = 1;
  google.protobuf.FloatValue float_value = 2;
  google.protobuf.Int32Value int32_value = 3;
  google.protobuf.Int64Value int64_value = 4;
  google.protobuf.Uint32Value uint32_value = 5;
  google.protobuf.Uint64Value uint64_value = 6;
  google.protobuf.BoolValue bool_value = 7;
  google.protobuf.BytesValue bytes_value = 8;
  google.protobuf.StringValue string_value = 9;
}
```

```typescript
type SampleMessage = Readonly<{
  doubleValue: number | undefined;
  floatValue: number | undefined;
  int32Value: number | undefined;
  int64Value: number | undefined;
  uint32Value: number | undefined;
  uint64Value: number | undefined;
  boolValue: boolean | undefined;
  bytesValue: string | undefined;
  stringValue: string | undefined;
}>;
```

## License

[MIT License](https://opensource.org/licenses/MIT).
