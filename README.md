# oasdiff-service

[![Go Report Card](https://goreportcard.com/badge/github.com/oasdiff/oasdiff-service)](https://goreportcard.com/report/github.com/oasdiff/oasdiff-service)

### Creating a tenant
Create a tenant and get a tenant ID:

curl -d '{"tenant": "my-company", "email": "james@my-company.com"}' https://auth.oasdiff.com/tenants
You will get a response with your tenant ID:

{"id": "2ahh9d6a-2221-41d7-bbc5-a950958345"}

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
