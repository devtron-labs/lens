Application to analyse git and relase data

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
