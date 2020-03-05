# Kubernetes CRD to OpenAPISchema

This project is a CLI that will convert a `kind: CustomResourceDefinition` to an OpenAPISchema JSON file that's compatible with [kubeval](https://kubeval.instrumenta.dev/).

## Motivation

As part of the Replicated vendor tools, we want to help ensure that valid YAML is created for every release. As more CRDs are used in applications, the schema has become more dynamic. Many projects publish CRDs without also publishing an OpenAPISchema for the project.

## Usage

```
crd-to-openapischema <url or path to schema>
```
