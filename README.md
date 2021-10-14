# GraphQL Introspection to SDL

Generates a graphql server schema using introspection.

Introspection query document mirrors the [graphql-js](https://github.com/graphql/graphql-js) `getIntrospectionQuery` document albeit compliant to the [June 2018 specification](https://spec.graphql.org/June2018/#sec-Introspection).

```sh
go run main.go > schema.graphql
```
