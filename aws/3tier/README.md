
aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --query "Vpcs[0].VpcId" --output text

aws ec2 describe-subnets --filters "Name=vpc-id,Values=VPCVALUE" --query "Subnets[*].[SubnetId,AvailabilityZone]" --output table

aws cloudformation validate-template --template-body file://template.yaml




echo "GET http://3tier-alb-1005396197.us-east-1.elb.amazonaws.com" | vegeta attack  -duration=5m | vegeta report



  aws ssm send-command \
    --instance-ids i-08a99eee93ac2b56b \
    --document-name "AWS-RunShellScript" \
    --parameters 'commands=["cat /var/log/cloud-init-output.log | tail -100"]' \
    --output text --query "Command.CommandId"


EC2 init script logs
cat /var/log/cloud-init-output.log