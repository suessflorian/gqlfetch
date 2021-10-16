# GQLFetch

GraphQL introspection based schema generator, introspection query document mirrors the [graphql-js](https://github.com/graphql/graphql-js) `getIntrospectionQuery` document albeit compliant to the [June 2018 specification](https://spec.graphql.org/June2018/#sec-Introspection).

## Usage

```go
import (
	"github.com/suessflorian/gqlfetch"
)

func main() {
  schema, _ := gqlfetch.BuildClientSchema(ctx, endpoint)
}
```

### Or use as cli tool

```bash
go get github.com/suessflorian/gqlfetch
gqlfetch --endpoint "localhost:8080/query" > schema.graphql
```

If you get an error claiming that `gqlfetch` cannot be found or is not defined, you may need to add `~/go/bin` to your `$PATH` (MacOS/Linux), or `%HOME%\go\bin` (Windows).
