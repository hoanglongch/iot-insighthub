# IoT InsightHub

## Overview
IoT InsightHub is a high-throughput telemetry and edge processing platform engineered for industrial environments. The system ingests and processes over 100K events per second from edge gateways using Amazon Kinesis, provides real‑time key performance indicators (KPIs) through Prometheus and Grafana, and supports sub‑25ms end‑to‑end latency. The platform also integrates advanced edge analytics using WebAssembly (compiled via TinyGo) for in‑browser anomaly detection and secure peer-to‑peer video and data sharing using WebRTC.

This project is built with a microservices architecture using Go for critical backend services, Kubernetes for orchestration, and Terraform for managing AWS infrastructure. CI/CD is implemented via GitHub Actions, and deployments are handled using GitOps principles with ArgoCD.

## Folder Structure
- **/cmd**  
  Contains main entry points for services:
  - `secure-api`: A Go service providing a secure API with JWT-based authentication and data persistence (backed by TimescaleDB).
  - `telemetry-ingestor`: A Go service that ingests telemetry events from Kinesis, processes them, and exposes Prometheus metrics.

- **/pkg**  
  Reusable libraries used by the services:
  - `api`: Shared API contracts and data structures.
  - `auth`: Authentication middleware and security utilities.
  - `kinesis`: Kinesis consumer implementations.
  - `secureapi`: Business logic for secure API operations (including robust database access with fault tolerance).
  - `telemetry`: Logic for processing telemetry events.

- **/wasm**  
  Contains the WebAssembly module for anomaly detection, written in Go (TinyGo). The module exports functions for anomaly detection and performance benchmarking.

- **/web**  
  Front-end assets and JavaScript code for:
  - Displaying local and remote video streams.
  - Integrating WebRTC signaling via a dedicated signaling server.
  - Loading and integrating the WASM module, with built-in telemetry and error reporting.

- **/deploy**  
  Contains Kubernetes manifests and ArgoCD deployment files (e.g., `secure-api-deploy.yaml` and `telemetry-ingestor-deploy.yaml`).

- **/infrastructure**  
  Terraform configuration files for provisioning AWS resources:
  - `main.tf`: Core resources (Kinesis, RDS with TimescaleDB, S3 for archive, Lambda, IAM).
  - `variables.tf`: Input variables.
  - `outputs.tf`: Resource outputs.
  - (Additional files such as `backend.tf` and module folders can be added to modularize the configuration.)

- **/github/workflows**  
  CI/CD pipeline configuration using GitHub Actions (e.g., `ci-cd.yml`) that covers linting, testing, Docker image building and scanning, Terraform deployment, and ArgoCD synchronization.

- **/tools**  
  Utility scripts, including:
  - `audit/audit.go`: A script to perform a baseline audit of your system (querying Prometheus metrics and testing API endpoints).

- **/docs**  
  Documentation assets including:
  - `DEPLOYMENT_WORKFLOW.md`: Describes deployment procedures and rollback strategies.
  - `ERROR_HANDLING.md`: Outlines common error scenarios and troubleshooting steps.
  - Additional documentation such as API contracts, architecture diagrams, and training materials.

## Prerequisites
Before you begin, ensure you have the following installed and configured:
- **Go** (version 1.18 or later)
- **TinyGo** (for compiling the WASM module)
- **Docker** (for building and running containers)
- **Terraform** (for AWS infrastructure provisioning)
- **AWS CLI** (configured with proper credentials)
- **kubectl** (for interacting with your Kubernetes cluster)
- **Helm** (for deploying observability and other charts)
- **ArgoCD CLI** (if using GitOps for Kubernetes deployments)
- **Git** (for version control)
- **k6/Locust/JMeter** (for load testing, as needed)

## Setup

### Local Development
1. **Clone the Repository:**
```bash
git clone <repository-url> cd InsightIoTHub
```

2. **Set Environment Variables:**
Create a `.env` file or configure your shell with the necessary environment variables (e.g., `JWT_SECRET`, database credentials). For example:
```bash
export JWT_SECRET="YourSuperSecretKey" export DB_USERNAME="your_db_username" export DB_PASSWORD="your_db_password" export BUCKET_NAME="your-archive-bucket" export TERRAFORM_STATE_BUCKET="your-state-bucket" export TERRAFORM_LOCK_TABLE="your-lock-table"
```

3. **Initialize the Go Module:**
```bash
go mod tidy
```

### Building the Services
- **Secure API:**
```bash
cd cmd/secure-api go build -o secure-api .
```
- **Telemetry Ingestor:**
```bash
cd cmd/telemetry-ingestor go build -o telemetry-ingestor .
```


### Compiling the WASM Module
From the project root or within the `/wasm` folder:
```bash
tinygo build -o anomaly_detection.wasm -target wasm ./wasm/anomaly_detection.go
```

Ensure the `anomaly_detection.wasm` file is placed in the `/web` directory for the front-end to load.

## Testing

### Unit and Integration Tests
Run Go tests for all modules:
```bash
go test ./...
```

The repository includes tests for:
- JWT authentication middleware
- API handlers for the secure API
- Database persistence (with simulated failure modes)
- Additional modules can be added as needed

### Load Testing
Use the k6 script located in `/tests/load/load_test.js` to simulate concurrent requests to your `/ingest` endpoint:
```bash
k6 run tests/load/load_test.js
```


## Deployment

### Infrastructure Deployment (Terraform)
1. **Initialize and Apply:**
   From the `/infrastructure` directory:
    ```bash
    terraform init terraform plan -out=tfplan terraform apply -auto-approve tfplan
    ```

2. **Remote State:**  
Ensure you have set up your remote state by creating the necessary S3 bucket and DynamoDB table (as described in `backend.tf` and `state-resources.tf` if available).

### Application Deployment (Kubernetes)
- **Deploy using ArgoCD:**  
Sync your application using:
```bash
argocd app sync iot-insighthub
```
- **Helm:**  
If using Helm charts, upgrade your deployments:
```bash
helm upgrade <release-name> ./chart-directory --namespace <namespace>
```
- **Verification:**  
Confirm rollouts with:
```bash
kubectl rollout status deployment/secure-api -n <namespace> kubectl rollout status deployment/telemetry-ingestor -n <namespace>
```


## Observability
- **Monitoring:**  
Deploy the Prometheus Operator using the kube-prometheus-stack Helm chart:
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts helm repo update helm install prometheus prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace
```

- **Log Aggregation:**  
Install Fluent Bit (or Fluentd) via Helm:
```bash
helm repo add fluent https://fluent.github.io/helm-charts helm repo update helm install fluent-bit fluent/fluent-bit --namespace logging --create-namespace
```

- **Distributed Tracing:**  
Deploy Jaeger for tracing:
```bash
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts helm repo update helm install jaeger jaegertracing/jaeger --namespace observability --create-namespace --set provisionDataStore.cassandra=false --set provisionDataStore.elasticsearch=false
```

Access Grafana via port-forwarding and use Jaeger UI to review trace data.

## CI/CD

The CI/CD pipeline is defined in `.github/workflows/ci-cd.yml` and performs:
- Linting, unit, and integration tests
- Docker image building (multi-stage builds), vulnerability scanning, and push to Docker repository
- Terraform deployment
- ArgoCD synchronization and Kubernetes rollout verification
- Artifact upload and changelog generation for auditing

## Documentation

Additional documentation is provided in the `/docs` folder:
- `DEPLOYMENT_WORKFLOW.md`: Details deployment procedures and rollback strategies.
- `ERROR_HANDLING.md`: Lists common errors with troubleshooting steps.
- `API_CONTRACTS.md`: Describes the API endpoints and data contracts.
- `TRAINING.md`: Contains training materials and scheduled training sessions.

## Contributing

Please review our contribution guidelines before submitting a pull request. All major changes should be accompanied by updates to tests and documentation. Follow the GitOps principles for infrastructure changes.

## License

[Your License Information Here]

## Contact

For questions or further guidance, please contact the project maintainers or check the issue tracker.























