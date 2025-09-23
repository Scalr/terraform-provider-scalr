import json
import boto3
import os

ecs = boto3.client('ecs')

def lambda_handler(event, context):
    print("Received event: " + json.dumps(event))
    headers = event.get("headers", {})
    print("Headers: ", headers)

    # Get subnet IDs from environment variable
    subnet_ids = os.environ['SUBNET_IDS'].split(',')

    try:
        ecs.run_task(
            cluster=os.environ['CLUSTER'],
            launchType='FARGATE',
            taskDefinition=os.environ['TASK_DEFINITION'],
            networkConfiguration={
                'awsvpcConfiguration': {
                    'subnets': subnet_ids,
                    'securityGroups': [os.environ['SECURITY_GROUP']],
                    'assignPublicIp': 'ENABLED'
                }
            }
        )
    except Exception as e:
        return {
            'statusCode': 500,
            'body': json.dumps(str(e))
        }

    return {
        'statusCode': 200,
        'body': json.dumps('Fargate task triggered!')
    }