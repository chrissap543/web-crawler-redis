#!/bin/bash
# scripts/connect-to-container.sh

# Get the running task
TASK_ARN=$(aws ecs list-tasks \
  --cluster scraper-cluster \
  --region us-west-2 \
  --desired-status RUNNING \
  --query 'taskArns[0]' \
  --output text)

if [ "$TASK_ARN" == "None" ] || [ -z "$TASK_ARN" ]; then
  echo "No running tasks found"
  exit 1
fi

TASK_ID=$(echo $TASK_ARN | cut -d'/' -f3)
echo "Connecting to task: $TASK_ID"

# Connect to scraper container
aws ecs execute-command \
  --cluster scraper-cluster \
  --task $TASK_ID \
  --container scraper \
  --command "/bin/sh" \
  --interactive \
  --region us-west-2
