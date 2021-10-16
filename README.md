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
Introduced a directory here `/gqlfetch` which will create a `gqlfetch` cli tool.

```bash
go install github.com/suessflorian/gqlfetch/gqlfetch@v1.0.0
gqlfetch --endpoint "localhost:8080/query" > schema.graphql
```

If you get an error claiming that `gqlfetch` cannot be found or is not defined, you may need to add `~/go/bin` to your `$PATH` (MacOS/Linux), or `%HOME%\go\bin` (Windows).

## Roadmap

Hoping to incorporate this into a graphql utility package like [gqlparser](https://github.com/vektah/gqlparser), enable introspection based type-safe graphql client generation in [genqlient](https://github.com/Khan/genqlient) and simplify the `Service` type federated services need to expose to a [bramble](https://movio.github.io/bramble/#/federation) gateway (improving the schema gathering logic).
