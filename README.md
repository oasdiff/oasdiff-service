# oasdiff-service

### Run diff on localhost
```
curl -X POST \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    http://localhost:8080/diff
```

### Run breaking-changes using cloud-run
```
curl -X POST \
    -F base=@data/openapi-test1.yaml \
    -F revision=@data/openapi-test3.yaml \
    https://api.oasdiff.com/breaking-changes
```
