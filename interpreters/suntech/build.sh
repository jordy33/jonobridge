go build
docker images --filter=reference="*suntechprotocol*" --format "{{.ID}}" | xargs docker rmi -f
cd /home/ubuntu/jonobridge/pkg/interpreters/suntechprotocol
docker build -t suntechprotocol -f ./Dockerfile .
docker tag suntechprotocol maddsystems/suntechprotocol:1.0.0
docker push maddsystems/suntechprotocol:1.0.0
