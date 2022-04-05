# Minion

This application is responsible to for create thumbnail for images and videos uploaded to minio ...

This is a "microservice" application intended to be part of a microservice architecture.

## Architecture

![Alt text](api-docs/images/minion.png "Architect")

## Development

To start your application in the dev profile, run:

```
go run cmd/main.go
```

## Building for production

### Packaging as executable file

To build the final runnable application, run:

```
go build -o minion cmd/main.go
```

## DevOps

### Health 

To check service health status use following link:

```
/api/v1/health
curl -XGET http://127.0.0.1:8080/api/v1/health
## response 
{
  "ok" : true
}
```
### Service info 
To check service info use following link:

```
/api/v1/info
curl -XGET http://127.0.0.1:8080/api/v1/info
##response
{
  "desc":"Thumbnailer for images and videos uploaded to minio",
  "name":"Minion",
  "tech-used":"go 1.17",
  "version":"1.0.0"
}
```

### Monitoring Metrics
```
/metrics
curl -XGET http://127.0.0.1:8080/metrics
```

### Functionality test
```

```

### Environment variables
```
##### Log
- MINION_LOG_LEVEL: 0

##### Server
- MINION_SERVER_ADDR: 127.0.0.1:8080
- MINION_SERVER_CERT: cert.pem
- MINION_SERVER_KEY: key.pem

##### Minio
- MINION_MINIO_ENDPOINT: 127.0.0.1
- MINION_MINIO_ACCESSKEYID: minion
- MINION_MINIO_SECRETACCESSKEY: minion
- MINION_MINIO_USESSL: false

##### Observability
- MINION_OBSERVABILITY_PROMETHEUS_ENABLED: true
```


