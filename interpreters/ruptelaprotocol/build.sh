go build
docker images --filter=reference="*ruptelaprotocol*" --format "{{.ID}}" | xargs docker rmi -f
cd /home/ubuntu/jonobridge/pkg/ruptelaprotocol
docker build -t ruptelaprotocol -f ./Dockerfile .
docker tag ruptelaprotocol maddsystems/ruptelaprotocol:1.0.0
docker push maddsystems/ruptelaprotocol:1.0.0