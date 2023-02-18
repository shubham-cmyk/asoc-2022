## Description
Controller-Design Contains the Architecure Flow of the controller.

User implements
1. Serverless CR
2. Secrets for Vendors
3. Function code 

The serverless CR is recognized by the controller and a job is created to perform s-deploy on the Mentioned Cloud Provider.The credential information is managed in the secrets are are loaded as the env variable during the execution of the job.

## Execution flow of a s-JOB
1. When a Serverless-CR is created a s-deploy is peformed and when Serverles-CR is deleted a s-remove is performed.
2. The Execution Type is managed by the Controller.
3. A general Job Created Volumes creates 3 Volumes that is mounted on the pod. 
4. One Managed Secret file for controlling the credentials of the Vendors.
5. The pod download the Docker Image from the docker-registry the image is build on node.js and the serverless binary comes pre-installed init. 
6. When Performing s-deploy the Pod Instruct the s binary to perform s-deploy and vice-versa.
7. After Performing the s-deploy the pod is deleted. 