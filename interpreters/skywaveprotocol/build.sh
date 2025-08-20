go build
docker images --filter=reference="*skywaveprotocol*" --format "{{.ID}}" | xargs docker rmi -f
cd /home/ubuntu/jonobridge/pkg/interpreters/skywaveprotocol
docker build -t skywaveprotocol -f ./Dockerfile .
docker tag skywaveprotocol maddsystems/skywaveprotocol:1.0.0
docker push maddsystems/skywaveprotocol:1.0.0
