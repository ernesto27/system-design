# 3-Tier Web Application on AWS

A CloudFormation-based 3-tier architecture with Auto Scaling EC2 instances, Application Load Balancer, and RDS PostgreSQL.

## Architecture

```
Internet → ALB → Auto Scaling Group (EC2) → RDS PostgreSQL
```

- **Web Tier**: EC2 instances running Docker containers behind an ALB
- **App Tier**: Bun.js application in Docker (pulled from ECR)
- **Data Tier**: RDS PostgreSQL database

## Prerequisites

### IAM Permissions

Your IAM user needs the following AWS managed policies:

| Policy | Purpose |
|--------|---------|
| `AmazonEC2FullAccess` | Launch Templates, Security Groups |
| `AutoScalingFullAccess` | Auto Scaling Groups, Scaling Policies |
| `ElasticLoadBalancingFullAccess` | ALB, Target Groups, Listeners |
| `AmazonRDSFullAccess` | RDS Database, Subnet Groups |
| `IAMFullAccess` | EC2 Roles, Instance Profiles |
| `AmazonSSMReadOnlyAccess` | Resolve AMI IDs from SSM parameters |

**To attach policies:**
1. IAM Console → Users → Select your user
2. Permissions → Add permissions → Attach policies directly
3. Search and attach each policy listed above

### Other Requirements

- AWS CLI configured (`aws configure`)
- Docker installed (for building images)
- An EC2 Key Pair created in your region

## Initial Deployment

### 1. Get VPC and Subnet IDs

```bash
# Get default VPC ID
aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --query "Vpcs[0].VpcId" --output text

# Get subnet IDs (replace VPC_ID)
aws ec2 describe-subnets --filters "Name=vpc-id,Values=VPC_ID" --query "Subnets[*].[SubnetId,AvailabilityZone]" --output table
```

### 2. Validate template

```bash
aws cloudformation validate-template --template-body file://template.yaml
```

### 3. Create stack

```bash
aws cloudformation create-stack \
  --stack-name 3tier \
  --template-body file://template.yaml \
  --parameters \
    ParameterKey=VpcId,ParameterValue=vpc-xxxxxxxx \
    ParameterKey=SubnetIds,ParameterValue="subnet-xxxxx,subnet-yyyyy" \
    ParameterKey=KeyPair,ParameterValue=your-keypair-name \
    ParameterKey=DBPassword,ParameterValue=your-secure-password \
    ParameterKey=ContainerImage,ParameterValue=383660184915.dkr.ecr.us-east-1.amazonaws.com/bunjs:v1 \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

## Deploying a New Version

### Step 1: Build and push new image

```bash
cd app

# Push with a version tag (recommended)
./push-ecr.sh v2

# Or use 'latest' (not recommended for production)
./push-ecr.sh
```

### Step 2: Update CloudFormation stack

```bash
aws cloudformation update-stack \
  --stack-name 3tier \
  --template-body file://template.yaml \
  --parameters \
    ParameterKey=ContainerImage,ParameterValue=383660184915.dkr.ecr.us-east-1.amazonaws.com/bunjs:v2 \
    ParameterKey=VpcId,UsePreviousValue=true \
    ParameterKey=SubnetIds,UsePreviousValue=true \
    ParameterKey=InstanceType,UsePreviousValue=true \
    ParameterKey=DBPassword,UsePreviousValue=true \
    ParameterKey=KeyPair,UsePreviousValue=true \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

### Step 3: Trigger Instance Refresh

After CloudFormation update completes, replace running instances with new ones:

**Via AWS Console:**
1. EC2 → Auto Scaling Groups → `3tier-web-asg`
2. Instance refresh tab → Start instance refresh
3. Set Minimum healthy percentage: 50%
4. Set Instance warmup: 120 seconds
5. Click Start

**Via CLI:**
```bash
aws autoscaling start-instance-refresh \
  --auto-scaling-group-name 3tier-web-asg \
  --preferences '{"MinHealthyPercentage": 50, "InstanceWarmup": 120}' \
  --region us-east-1
```

### Step 4: Monitor deployment

```bash
# Check instance refresh status
aws autoscaling describe-instance-refreshes \
  --auto-scaling-group-name 3tier-web-asg \
  --region us-east-1

# Check CloudFormation stack status
aws cloudformation describe-stacks \
  --stack-name 3tier \
  --query 'Stacks[0].StackStatus'
```


### Load testing with Vegeta
```bash
echo "GET http://3tier-alb-xxxxx.us-east-1.elb.amazonaws.com" | vegeta attack -duration=5m | vegeta report
```

### View EC2 init logs (via SSM)
```bash
aws ssm send-command \
  --instance-ids i-xxxxxxxxx \
  --document-name "AWS-RunShellScript" \
  --parameters 'commands=["cat /var/log/cloud-init-output.log | tail -100"]' \
  --output text --query "Command.CommandId"
```


