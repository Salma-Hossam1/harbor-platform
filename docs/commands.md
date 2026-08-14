## docker

1-  docker compose config
2- docker compose up -d
3- docker compose down -v 
4- docker compose down
5- docker compose logs
docker compose ps
docker network inspect <network_name>

check container ready for postgres :
docker exec -it harbor-postgres pg_isready -U harbor
check redis :
docker exec -it harbor-redis redis-cli ping



# k8s

 kubectl get endpointslice -n artifact-platform
port forward the clusterIP sevice : kubectl port-forward -n artifact-platform svc/event-gateway 8080:8080