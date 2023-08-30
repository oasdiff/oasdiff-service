# oasdiff-service

[![Go Report Card](https://goreportcard.com/badge/github.com/oasdiff/oasdiff-service)](https://goreportcard.com/report/github.com/oasdiff/oasdiff-service)

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
