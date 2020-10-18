# private files: ca.key, server.key, server.pem, server.crt
#  share files: server.pem (needed by the CA), ca.crt (needed by client)


# server name
SERVER_CN=localhost

# Step1: Generate Certificate Authority + Trust Certificate (ca.crt)
openssl genrsa -passout pass:1111 -des3 -out ca.key 4096
openssl req -passin pass:1111 -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=${SERVER_CN}"

# Step2: Generate the server private key server.key
openssl genrsa -passout pass:1111 -des3 -out server.key 4096

# Step3: Get a certificate Signing request from the CA server.csr
openssl req -passin pass:1111 -new -key server.key -out server.csr -subj "/CN=${SERVER_CN}"

# Step4: Sign the certificate with the CA we created (id called self signing)
openssl x509 -req -passin pass:1111 -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt

# Step5: Convert the server certificate to .pem format (server.pem) - usable by gRPC
openssl pkcs8 -topk8 -nocrypt -passin pass:1111 -in server.key -out server.pem




