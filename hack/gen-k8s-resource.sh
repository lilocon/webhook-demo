source .env

cat << EOF > .run/server-csr.json
{
    "CN": "${APP}",
    "hosts": [
        "localhost",
        "${APP}",
        "${APP}.${NAMESPACE}",
        "${APP}.${NAMESPACE}.svc",
        "${APP}.${NAMESPACE}.svc.cluster",
        "${APP}.${NAMESPACE}.svc.cluster.local"
    ],
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [{
        "C": "CN",
        "ST": "BJ",
        "L": "BJ",
        "O": "Kubernetes",
        "OU": "Kubernetes-manual"
    }]
}
EOF

cfssl gencert \
  -ca=./ca/ca.pem \
  -ca-key=./ca/ca-key.pem \
  -profile=kubernetes \
  .run/server-csr.json | cfssljson -bare .run/server

CA_BUNDLE=$(cat ca/ca.pem | base64 | tr -d '\n')
CERT_PEM=$(cat .run/server.pem | base64 | tr -d '\n')
KEY_PEM=$(cat .run/server-key.pem | base64 | tr -d '\n')

rm .run/server.csr
rm .run/server-csr.json
rm .run/server.pem
rm .run/server-key.pem

cp resource.yaml .run/resource.yaml

sed -i "" "s@\${APP}@${APP}@" .run/resource.yaml
sed -i "" "s@\${NAMESPACE}@${NAMESPACE}@" .run/resource.yaml
sed -i "" "s@\${IMAGE}@${IMAGE}@" .run/resource.yaml
sed -i "" "s@\${VERSION}@${VERSION}@" .run/resource.yaml
sed -i "" "s@\${CA_BUNDLE}@${CA_BUNDLE}@" .run/resource.yaml
sed -i "" "s@\${CERT_PEM}@${CERT_PEM}@" .run/resource.yaml
sed -i "" "s@\${KEY_PEM}@${KEY_PEM}@" .run/resource.yaml
sed -i "" "s@\${NAMESPACE_MATCH_LABEL}@${NAMESPACE_MATCH_LABEL}@" .run/resource.yaml