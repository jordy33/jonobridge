go build
docker images --filter=reference="*huabaoprotocol*" --format "{{.ID}}" | xargs docker rmi -f
cd /home/ubuntu/jonobridge/pkg/interpreters/huabaoprotocol
docker build -t huabaoprotocol -f ./Dockerfile .
docker tag huabaoprotocol maddsystems/huabaoprotocol:1.0.0
docker push maddsystems/huabaoprotocol:1.0.0
