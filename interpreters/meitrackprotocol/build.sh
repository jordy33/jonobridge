go build
docker images --filter=reference="*meitrackprotocol*" --format "{{.ID}}" | xargs docker rmi -f
cd /home/ubuntu/jonobridge/pkg/interpreters/meitrackprotocol
docker build -t meitrackprotocol -f ./Dockerfile .
docker tag meitrackprotocol maddsystems/meitrackprotocol:1.0.0
docker push maddsystems/meitrackprotocol:1.0.0
