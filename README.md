# crdtohttp

A simple CRD that when created initiates a HTTP request. This resource creates a one off HTTP request and is immutable.


```
cat << EOF | k apply -f -
apiVersion: openapi.pivotal.io/v1
kind: Request
metadata:
  name: request-sample
spec:
  path: http://dummy.restapiexample.com/api/v1/create
  method: post
  body: '{"name":"test","salary":"123","age":"23"}'
  headers:
    - "Content-type: application/json"
EOF
```

Succeeds with:
```
apiVersion: openapi.pivotal.io/v1
kind: Request
metadata:
  name: request-sample
  namespace: 
spec:
  body: '{"name":"test","salary":"123","age":"23"}'
  headers:
  - 'Content-type: application/json'
  method: post
  path: http://dummy.restapiexample.com/api/v1/create
status:
  body: '{"status":"success","data":{"name":"test","salary":"123","age":"23","id":39}}'
  code: "200"
```
