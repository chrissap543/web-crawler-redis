#!/usr/bin/env sh

set -e

source ./aws-config.env

export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Check if variables are loaded
echo "AWS_ACCOUNT_ID: $AWS_ACCOUNT_ID"
echo "SUBNET1_ID: $SUBNET1_ID"
echo "SUBNET2_ID: $SUBNET2_ID"
echo "SECURITY_GROUP_ID: $SECURITY_GROUP_ID"

aws iam get-role --role-name ecsTaskExecutionRole 2>/dev/null || {
  echo "Creating ecsTaskExecutionRole..."

  cat > /tmp/trust-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

 aws iam create-role \
    --role-name ecsTaskExecutionRole \
    --assume-role-policy-document file:///tmp/trust-policy.json

  aws iam attach-role-policy \
    --role-name ecsTaskExecutionRole \
    --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy

  rm /tmp/trust-policy.json
}

envsubst < deployments/ecs/docker-compose.yml.template > deployments/ecs/docker-compose.yml
envsubst < deployments/ecs/ecs-params.yml.template > deployments/ecs/ecs-params.yml

cd deployments/ecs
ecs-cli compose \
    --project-name web-scraper \
    --file docker-compose.yml \
    --ecs-params ecs-params.yml \
    service down \
    --cluster-config scraper-config \
    --ecs-profile default 2 >/dev/null || echo "No existing service to stop"

echo "deploying new"
ecs-cli compose \
  --project-name web-scraper \
  --file docker-compose.yml \
  --ecs-params ecs-params.yml \
  service up \
  --cluster-config scraper-config \
  --ecs-profile default \
  --create-log-groups
cd ../..
