# Load Tests

k6 load test scripts for Myrmex production staging.

## Prerequisites

Install k6: https://k6.io/docs/get-started/installation/

```bash
# macOS
brew install k6

# Docker
docker pull grafana/k6
```

## Usage

```bash
export BASE_URL=https://your-core.run.app
export AUTH_TOKEN=your-jwt-token

# Login storm (100 VUs, 2 min)
k6 run deploy/load-tests/auth-flow.js

# API CRUD mix (200 VUs, 5 min)
k6 run deploy/load-tests/api-crud.js

# Mixed workload ramp (0→500 VUs)
k6 run deploy/load-tests/mixed-workload.js
```

## Targets

| Test         | VUs  | Duration | p95 latency | Error rate |
|--------------|------|----------|-------------|------------|
| auth-flow    | 100  | 2m       | < 500ms     | < 1%       |
| api-crud     | 200  | 5m       | < 300ms     | < 1%       |
| mixed        | 500  | 10m ramp | < 1000ms    | < 1%       |

## Notes

- `BASE_URL` defaults to `https://core-placeholder.run.app` if not set
- `AUTH_TOKEN` must be a valid JWT for authenticated endpoints
- Test users `test-N@hcmus.edu.vn` must exist in staging DB for auth-flow
