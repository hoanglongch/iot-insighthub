# DEPLOYMENT_WORKFLOW.md

## Overview
This document outlines the procedures for deploying the IoT InsightHub platform. It covers both the provisioning of infrastructure via Terraform and the deployment of application services on Kubernetes using ArgoCD and Helm. The document also details rollback procedures and post-deployment verification steps.

---

## 1. Infrastructure Deployment (Terraform)

### Preparation
- Ensure that all infrastructure changes (Terraform modules, backend configuration, etc.) are fully reviewed and merged into the main branch.
- Confirm that any environment-specific variables (such as DB credentials, bucket names, etc.) are updated in your terraform.tfvars file or provided via environment variables.

### Deployment Steps
1. **Initialize Remote State:**
   Run the following command from the infrastructure directory:
   
   terraform init

2. **Plan the Changes:**
   Generate and review an execution plan:
   
   terraform plan -out=tfplan

3. **Apply the Changes:**
   Execute the planned changes:
   
   terraform apply -auto-approve tfplan

4. **Post-Apply Verification:**
   - Verify that all new and updated resources are deployed in AWS (check S3, RDS, Kinesis, etc.).
   - Use the outputs (such as rds_endpoint and kinesis_stream_name) to confirm connectivity and configuration.

---

## 2. Application Deployment (Kubernetes via ArgoCD/Helm)

### Build and Push Docker Images
- Your CI/CD pipeline builds your Docker images via multi-stage Dockerfiles and pushes them to your Docker registry.
- Verify that the images include all necessary changes for your secure API and telemetry ingestor services.

### Deployment Steps
1. **Sync with ArgoCD:**
   Run the following command to synchronize your GitOps application:
   
   argocd app sync iot-insighthub

2. **Helm Upgrade (if applicable):**
   If deploying via Helm, upgrade your release as follows:
   
   helm upgrade <release-name> ./chart-directory --namespace <namespace>

3. **Verify Rollout:**
   Check that your deployments have rolled out correctly using:
   
   kubectl rollout status deployment/secure-api -n <namespace>
   kubectl rollout status deployment/telemetry-ingestor -n <namespace>

---

## 3. Rollback Procedures

### Application Rollback
- If a deployment fails, rollback using Kubernetes:
  
  kubectl rollout undo deployment/<deployment-name> -n <namespace>
  
- For Helm-managed deployments:
  
  helm rollback <release-name> <revision> --namespace <namespace>

### Infrastructure Rollback
- If an infrastructure change results in issues, revert the changes in version control and reapply Terraform:
  
  1. Revert the commit in Git.
  2. Run:
     
     terraform plan -out=tfplan
     terraform apply -auto-approve tfplan

---

## 4. Verification Post Deployment
- Validate that all pods are running and passing readiness and liveness probes.
- Monitor system metrics via Prometheus and Grafana dashboards.
- Review logs and distributed traces to confirm healthy operation.
- Document any post-deployment issues and follow up in review meetings.

