---
title: Using Multiple OLAP Engines
description: Using multiple OLAP engines to power dashboards in the same project
sidebar_label: Using Multiple OLAP engines
sidebar_position: 4
---

## Overview

If you have access to another OLAP engine (such as [ClickHouse](clickhouse.md) or [Druid](druid.md)), you have the option to either:
- Create dedicated projects that are powered by one specific OLAP engine (default)
- Use different OLAP engines _in the same project_ to power separate dashboards

On this page, we will walk through how to configure the latter. 