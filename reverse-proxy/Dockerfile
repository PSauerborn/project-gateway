FROM opentracing/nginx-opentracing:latest

COPY ./default.conf /etc/nginx/conf.d/default.conf
COPY ./nginx.conf /etc/nginx/nginx.conf
COPY ./jaeger-nginx-config.json /etc/jaeger-nginx-config.json
COPY ./sites-available /etc/nginx/conf.d/

COPY ./certs /etc/nginx/certs
COPY ./private-key.txt /etc/nginx/