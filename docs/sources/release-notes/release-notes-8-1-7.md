+++
aliases = ["/docs/grafana/latest/release-notes/release-notes-8-1-7/"]
title = "Release notes for Grafana 8.1.7"

[_build]
  list = false
+++

<!-- Auto generated by update changelog github action -->

# Release notes for Grafana 8.1.7

### Bug fixes

- **Alerting:** Fix alerts with evaluation interval more than 30 seconds resolving before notification. [#39513](https://github.com/grafana/grafana/pull/39513), [@gerobinson](https://github.com/gerobinson)
- **Elasticsearch/Prometheus:** Fix usage of proper SigV4 service namespace. [#39439](https://github.com/grafana/grafana/pull/39439), [@marefr](https://github.com/marefr)