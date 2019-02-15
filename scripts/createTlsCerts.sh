#!/bin/bash
# call this script with an email address (valid or not).
# like:
# ./makecert.sh joe@random.com
mkdir ../measurement-plane/tcp-tls-throughput/certs
rm ../measurement-plane/tcp-tls-throughput/certs/*
echo "make server cert"
openssl req -new -nodes -x509 -out ../measurement-plane/tcp-tls-throughput/certs/server.pem -keyout ../measurement-plane/tcp-tls-throughput/certs/server.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=www.random.com/emailAddress=$1"
echo "make client cert"
openssl req -new -nodes -x509 -out ../measurement-plane/tcp-tls-throughput/certs/client.pem -keyout ../measurement-plane/tcp-tls-throughput/certs/client.key -days 3650 -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=www.random.com/emailAddress=$1"