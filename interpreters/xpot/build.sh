go build
docker images --filter=reference="*xpot*" --format "{{.ID}}" | xargs docker rmi -f
cd /home/ubuntu/jonobridge/pkg/xpot
docker build -t xpot -f ./Dockerfile .
docker tag xpot maddsystems/xpot:1.0.0
docker push maddsystems/xpot:1.0.0
