# Runbook: Secure API Service Outage

## Incident Summary
- **Time Detected:** [Time]
- **Impact:** [Service affected, e.g., "Ingestion API not responding"]
- **Reported by:** [Monitoring alert or user report]

## Step 1: Initial Assessment
- Verify service health by checking Kubernetes pod status:
  ```bash
  kubectl get pods -n <namespace> -l app=secure-api
  ```
- Review logs of the affected pods:
  ```bash
  kubectl logs <pod-name> -n <namespace>
  ```

## Step 2: Mitigation
- If pods are crash-looping, consider restarting:
  ```bash
  kubectl rollout restart deployment/secure-api -n <namespace>
  ```
- If the error is due to database connection issues, verify RDS connectivity and alert the DB team.

## Step 3: Post-Incident
- Document the incident in the incident report database.
- Review metrics in Prometheus/Grafana.
- Schedule a post-mortem meeting for further analysis and long-term fixes.
