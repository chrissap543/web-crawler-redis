#!/bin/bash
# Neo4j Community Edition Setup Script - Using Official AMI

set -e

echo "🚀 Setting up Neo4j Community Edition using official AMI..."

# Variables
REGION="us-west-2"
INSTANCE_TYPE="t3.medium"
# KEY_NAME="your-key-pair"  # Replace with your key pair name
SECURITY_GROUP_NAME="neo4j-community-sg"

# Function to get the latest Neo4j Community Edition AMI
get_bitnami_ami() {
    # Search for Neo4j Community Edition AMI by known patterns
    aws ec2 describe-images \
        --region $REGION \
        --owners 679593333241 \
        --filters "Name=name,Values=*bitnami-neo4j*" \
                  "Name=state,Values=available" \
        --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' \
        --output text
}
AMI_ID=$(get_bitnami_ami)

aws ec2 create-security-group \
    --group-name $SECURITY_GROUP_NAME \
    --description "Security group for Neo4j Community Edition (Bitnami)" \
    --region $REGION 2>/dev/null || echo "Security group already exists"

# Get security group ID
SG_ID=$(aws ec2 describe-security-groups \
    --group-names $SECURITY_GROUP_NAME \
    --region $REGION \
    --query 'SecurityGroups[0].GroupId' \
    --output text)

# Neo4j HTTP port (Browser interface)
aws ec2 authorize-security-group-ingress \
    --group-id $SG_ID \
    --protocol tcp \
    --port 7474 \
    --cidr 0.0.0.0/0 \
    --region $REGION 2>/dev/null || echo "Port 7474 rule already exists"

# Neo4j Bolt port (Driver connections)
aws ec2 authorize-security-group-ingress \
    --group-id $SG_ID \
    --protocol tcp \
    --port 7687 \
    --cidr 0.0.0.0/0 \
    --region $REGION 2>/dev/null || echo "Port 7687 rule already exists"

# SSH access
aws ec2 authorize-security-group-ingress \
    --group-id $SG_ID \
    --protocol tcp \
    --port 22 \
    --cidr 0.0.0.0/0 \
    --region $REGION 2>/dev/null || echo "SSH rule already exists"

USER_DATA=$(cat << 'EOF'
#!/bin/bash
# Official Neo4j Community AMI setup
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)

# Log connection details
echo "=== Neo4j Community Edition Ready ===" > /var/log/neo4j-setup.log
echo "Browser URL: http://$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4):7474" >> /var/log/neo4j-setup.log
echo "Bolt URL: bolt://$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4):7687" >> /var/log/neo4j-setup.log
echo "Username: neo4j" >> /var/log/neo4j-setup.log
echo "Password: $INSTANCE_ID (your EC2 instance ID)" >> /var/log/neo4j-setup.log

# Ensure Neo4j is configured for external access (AMI should have this already)
# But let's make sure the service is running
systemctl enable neo4j 2>/dev/null || true
systemctl start neo4j 2>/dev/null || true
EOF
)

INSTANCE_ID=$(aws ec2 run-instances \
    --image-id $AMI_ID \
    --count 1 \
    --instance-type $INSTANCE_TYPE \
    --security-group-ids $SG_ID \
    --region $REGION \
    --user-data "$USER_DATA" \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=Neo4j-Community-Database}]' \
    --query 'Instances[0].InstanceId' \
    --output text)

echo "⏳ Waiting for instance to be running..."
aws ec2 wait instance-running --instance-ids $INSTANCE_ID --region $REGION

# Get public IP
PUBLIC_IP=$(aws ec2 describe-instances \
    --instance-ids $INSTANCE_ID \
    --region $REGION \
    --query 'Reservations[0].Instances[0].PublicIpAddress' \
    --output text)

echo ""
echo "🎉 Neo4j Community Edition instance is ready!"
echo "=================================================="
echo "Instance ID: $INSTANCE_ID"
echo "Public IP: $PUBLIC_IP"
echo "Browser URL: http://$PUBLIC_IP:7474"
echo "Bolt URL: bolt://$PUBLIC_IP:7687"
echo ""
echo "🔐 Login Credentials:"
echo "Username: neo4j"
echo "Password: $INSTANCE_ID (your EC2 instance ID)"
echo ""
echo "🔧 Connection details for your Go scraper:"
echo "NEO4J_URI=bolt://$PUBLIC_IP:7687"
echo "NEO4J_USER=neo4j"
echo "NEO4J_PASSWORD=$INSTANCE_ID"
echo ""
echo "📝 Important Notes:"
echo "- The password is your EC2 instance ID: $INSTANCE_ID"
echo "- On first login, you may be prompted to change the password"
echo "- Access the browser interface at: http://$PUBLIC_IP:7474"
