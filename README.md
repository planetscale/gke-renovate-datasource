# GKE Renovate Datasource

A [Renovate customDatasource](renovate-custom-datasource) that tracks Google Kubernetes Engine (GKE) versions across release channels.

## Features

- Scrapes official GKE release channel RSS feeds
- Provides version and release timestamp information (useful for Renovate's `minimumReleaseAge` option)
- Updated daily via GitHub Actions

## Available Channels

- rapid: https://raw.githubusercontent.com/planetscale/gke-renovate-datasource/main/static/rapid.json
- regular: https://raw.githubusercontent.com/planetscale/gke-renovate-datasource/main/static/regular.json
- stable: https://raw.githubusercontent.com/planetscale/gke-renovate-datasource/main/static/stable.json

## Usage

Add to your renovate.json:

```json
{
  "customDatasources": {
    "gke-stable": {
      "defaultRegistryUrlTemplate": "https://raw.githubusercontent.com/planetscale/gke-renovate-datasource/main/static/stable.json",
      "format": "json"
    }
  }
}
```
<!-- refs -->
[renovate-custom-datasource]: https://docs.renovatebot.com/modules/datasource/custom/