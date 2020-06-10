# protoc-gen-tstype

Generate TypeScript Readonly type from Proto file.

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
type SampleMessage = Readonly<{
}> &
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

**Implementation plan**

### Nested Types

**Not Support**

## License

[MIT License](https://opensource.org/licenses/MIT).
