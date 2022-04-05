package config

const Namespace = "minion"

const Default = `
log:
# 0 - INFO 1 - DEBUG 2 - TRACE
  level: 0

server:
  addr: ":8080"
  cert: "cert.pem"
  key: "key.pem"  

minio:
  addr: "127.0.0.1"
  accessKeyID: "minion"
  secretAccessKey: "minion"
  useSSL: false

thumbnailer:
  width: 200
  height: 200
  
observability:
  prometheus:
    enabled: true  
`
