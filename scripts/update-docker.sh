#!/usr/bin/env sh

AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
sudo docker build --no-cache -f deployments/docker/AWS_Dockerfile -t scraper:latest .
sudo docker tag scraper:latest $AWS_ACCOUNT_ID.dkr.ecr.us-west-2.amazonaws.com/web-scraper:latest
sudo docker push $AWS_ACCOUNT_ID.dkr.ecr.us-west-2.amazonaws.com/web-scraper:latest
