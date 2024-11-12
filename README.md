# ⚠️ Deprecated Repository

This repository has been deprecated and is no longer actively maintained.

## Deprecation Reason

This repository has been merged into our monorepo structure to reduce maintenance overhead.

## New Location

The code has been moved to:
[Lens](https://github.com/devtron-labs/devtron-services/tree/main/lens)

Please update your references and direct all future contributions, issues, and pull requests to the new repository location.

For any questions or concerns, please open an issue in the new repository.

Thank you for your understanding.

## Application to analyse git and relase data

### Initial testing the application
```bash
curl -XPOST localhost:8080/new-deployment-event -d '{
  "ApplicationId": 7,
  "EnvironmentId": 1,
  "CiArtifactId": 9,
  "ReleaseId": 16,
  "PipelineOverrideId": 102,
  "TriggerTime": "2019-10-28T16:41:13.698823+05:30",
  "PipelineMaterials": [
    {
      "PipelineMaterialId": 9,
      "CommitHash": "5e1bcba61adc9ee5826b56e0c3491d4aa023af33"
    }
  ]
}'
```

change `CiArtifactId `,  `ReleaseId`, `TriggerTime` and `CommitHash` for each release
