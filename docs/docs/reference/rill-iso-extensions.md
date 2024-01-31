---
title: Rill extensions to ISO 8601
description: All the extensions options to the standard ISO 8601 duration standard
sidebar_label: Rill extensions to ISO 8601
sidebar_position: 13
---

We have extended the ISO 8601 standard to specify ranges like `Week-to-Date`.

### Extensions

| Rill ISO extension | Description      | As Time Range | As Time Comparison |
|--------------------|------------------|---------------|--------------------|
| inf                | All time         | Yes           | No                 |
| rill-TD            | Today            | Yes           | No                 |
| rill-WTD           | Week to Date     | Yes           | No                 |
| rill-MTD           | Month to Date    | Yes           | No                 |
| rill-QTD           | Quarter to Date  | Yes           | No                 |
| rill-YTD           | Year to Date     | Yes           | No                 |
| rill-PP            | Previous Period  | No            | Yes                |
| rill-PD            | Previous Day     | No            | Yes                |
| rill-PW            | Previous Week    | No            | Yes                |
| rill-PM            | Previous Month   | No            | Yes                |
| rill-PQ            | Previous Quarter | No            | Yes                |
| rill-PY            | Previous Year    | No            | Yes                |

