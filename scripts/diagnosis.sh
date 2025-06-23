#!/bin/bash
echo "=== ECS RESOURCES ==="
echo "Clusters:"
aws ecs list-clusters --region us-west-2 --query 'clusterArns[]' --output table

echo -e "\nServices in scraper-cluster:"
aws ecs list-services --cluster scraper-cluster --region us-west-2 --query 'serviceArns[]' --output table 2>/dev/null || echo "No scraper-cluster found"

echo -e "\nTask Definitions:"
aws ecs list-task-definitions --region us-west-2 --query 'taskDefinitionArns[]' --output table | grep -E "(scraper|web-scraper)" || echo "No scraper task definitions found"

echo -e "\n=== ECR REPOSITORIES ==="
aws ecr describe-repositories --region us-west-2 --query 'repositories[].repositoryName' --output table | grep -E "(scraper|web-scraper)" || echo "No scraper repositories found"

echo -e "\n=== LOCAL DOCKER IMAGES ==="
docker images | head -1  # header
docker images | grep -E "(scraper|web-scraper)" || echo "No local scraper images found"

echo -e "\n=== ECS CLI CONFIG ==="
ecs-cli configure list
