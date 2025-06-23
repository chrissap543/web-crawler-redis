#!/usr/bin/env sh

# Get the VPC created by ecs-cli
VPC_ID=$(aws ec2 describe-vpcs \
  --filters "Name=tag:aws:cloudformation:stack-name,Values=amazon-ecs-cli-setup-scraper-cluster" \
  --query 'Vpcs[0].VpcId' --output text)

# Get the default security group for this VPC
SECURITY_GROUP_ID=$(aws ec2 describe-security-groups \
  --filters "Name=vpc-id,Values=$VPC_ID" "Name=group-name,Values=default" \
  --query 'SecurityGroups[0].GroupId' --output text)

echo "VPC ID: $VPC_ID"
echo "Default Security Group: $SECURITY_GROUP_ID"

# Add required security group rules
echo "Adding security group rules to default group..."

# Allow HTTP traffic on port 8080
aws ec2 authorize-security-group-ingress \
  --group-id $SECURITY_GROUP_ID \
  --protocol tcp \
  --port 8080 \
  --cidr 0.0.0.0/0

# Allow all traffic within security group (for Redis communication between containers)
aws ec2 authorize-security-group-ingress \
  --group-id $SECURITY_GROUP_ID \
  --protocol all \
  --source-group $SECURITY_GROUP_ID

echo "Security group rules added to default group!"

# Also get subnet information for later use
SUBNETS=$(aws ec2 describe-subnets \
  --filters "Name=vpc-id,Values=$VPC_ID" \
  --query 'Subnets[].SubnetId' --output text)

SUBNET_ARRAY=($SUBNETS)
SUBNET1=${SUBNET_ARRAY[0]}
SUBNET2=${SUBNET_ARRAY[1]}

echo "Subnets found: $SUBNET1, $SUBNET2"

# Save configuration for later scripts
cat > aws-config.env << EOF
VPC_ID=$VPC_ID
SUBNET1_ID=$SUBNET1
SUBNET2_ID=$SUBNET2
SECURITY_GROUP_ID=$SECURITY_GROUP_ID
REGION=us-west-2
CLUSTER_NAME=scraper-cluster
EOF

echo "✅ Configuration saved to aws-config.env"
