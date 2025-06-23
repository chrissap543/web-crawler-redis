#!/bin/bash
# Create ECR repository for your web scraper

REGION="us-west-2"
REPO_NAME="web-scraper"

echo "Creating ECR repository: $REPO_NAME"

# Create the repository
aws ecr create-repository \
  --region $REGION \
  --repository-name $REPO_NAME \
  --image-scanning-configuration scanOnPush=true \
  --encryption-configuration encryptionType=AES256

echo "ECR repository created!"

# Get your account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "📋 Repository Details:"
echo "Repository Name: $REPO_NAME"
echo "Repository URI: $ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/$REPO_NAME"
echo "Account ID: $ACCOUNT_ID"
echo "Region: $REGION"
