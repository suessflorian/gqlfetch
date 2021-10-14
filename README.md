# GraphQL Introspection to SDL

Generates a graphql server schema using introspection.

Introspection query document mirrors the [graphql-js](https://github.com/graphql/graphql-js) `getIntrospectionQuery` document albeit compliant to the [June 2018 specification](https://spec.graphql.org/June2018/#sec-Introspection).

```sh
go run main.go > schema.graphql
```

Far from complete:
- [ ] **Completely** derive the schema
- [ ] Rely on https://github.com/graphql-go/graphql schema types
- [ ] Black box testing input/outputs, matching that of [this gist using js spec reference implementation](https://gist.github.com/kyledetella/c671ca6335fbfd9e6aa3db97db0c212f)
