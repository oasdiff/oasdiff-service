# oasdiff-service

[![Go Report Card](https://goreportcard.com/badge/github.com/oasdiff/oasdiff-service)](https://goreportcard.com/report/github.com/oasdiff/oasdiff-service)

### Creating a tenant
Create a tenant and get a tenant ID:
```
curl -d '{"tenant": "my-company", "email": "james@my-company.com"}' https://register.oasdiff.com/tenants
```
You will get a response with your tenant ID:
```
{"id": "2ahh9d6a-2221-41d7-bbc5-a950958345"}
```
### Run diff on localhost
```
curl -X POST \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    http://localhost:8080/tenants/{tenant-id}/diff
```

### Run breaking-changes using cloud-run
```
curl -X POST \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/breaking-changes
```

### Run changelog using cloud-run
```
curl -X POST \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/changelog
```

### Response Formats
You can request the response as json:
```
curl -X POST -H "Accept: application/json" \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/breaking-changes
```

Or as yaml:
```
curl -X POST -H "Accept: application/yaml" \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/breaking-changes
```

Or as html:
```
curl -X POST -H "Accept: text/html" \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/breaking-changes
```

Or as text:
```
curl -X POST -H "Accept: text/plain" \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/breaking-changes
```

Or as markdown:
```
curl -X POST -H "Accept: text/markdown" \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/breaking-changes
```

### Output Languages
You can specify the output language using the `Accept-Language` header. Supported languages:
- `en` - English (default)
- `ru` - Russian  
- `pt-br` - Portuguese (Brazil)
- `es` - Spanish

Example with Spanish output:
```
curl -X POST -H "Accept: application/json" -H "Accept-Language: es" \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/breaking-changes
```

Example with Russian output:
```
curl -X POST -H "Accept: text/html" -H "Accept-Language: ru" \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/changelog
```

You can also specify language preferences with quality values:
```
curl -X POST -H "Accept-Language: pt-BR,pt;q=0.9,en;q=0.8" \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/tenants/{tenant-id}/diff
```

### Errors
oasdiff-service uses conventional HTTP response codes to indicate the success or failure of an API request. In general: Codes in the 2xx range indicate success. Codes in the 4xx range indicate a failure with additional information provided (e.g., invalid OpenAPI spec format, a required parameter was missing, etc.). Codes in the 5xx range indicate an error with oasdiff-service servers (these are rare)
