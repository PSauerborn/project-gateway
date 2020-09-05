# Introduction

Repository containing the infrastructure code required for the https://project-gateway.app
application, including the following components

1. `Reverse Proxy` - NGINX container used to control ingress and house TLS certs
2. `API Gateway` - NGINX container used to control ingress to micro-services
3. `CDN` - Frontend for application selection
