docker rm $(docker ps -aq)
docker volume rm -f $(docker volume ls -qf dangling=true | xargs)
make dev-docker
docker-compose -f demo/docker-compose-cluster/docker-compose.yml up
