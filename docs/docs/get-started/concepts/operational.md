---
title: Operational BI vs. Traditional BI
sidebar_label: What is Operational BI?
sidebar_position: 10
hide_table_of_contents: true
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Operational vs. Traditional BI

The distinction between operational and business intelligence is analogous to the distinction between fast and slow thinking, as characterized by the psychologist Daniel Kahneman in his book __Thinking, Fast and Slow__. One system operates quickly and automatically for simple decisions, and the other leverages slow and effortful deliberation for complex decisions. 

Ultimately, the output of both operational and business intelligence is decisions. Operational intelligence fuels fast, frequent decisions on real-time and near-time data by hands-on operators. Business intelligence drives complex decisions that occur daily or weekly, on fairly complete data sets. 

![operationalcomparison](/img/concepts/operational/comparison.png)

## Why Operational BI requires new tools

Operational intelligence provides a set of decision-making capabilities that are complementary to business intelligence, but its unique performance requirements also demand a novel stack of distinct technologies which are complementary and sit adjacent to existing business intelligence stacks.

Analytics technology stacks can be thought of as data flowing into a three-layered cake consisting of ETL, databases, and applications. The requirements for an operational intelligence stack are that it supports:

- high speed of data from ETL to application
- high frequency, low-latency queries between the database and application layer

In the diagram below we illustrate two common examples for technologies used in operational and business intelligence stacks.

![operationalbi](/img/concepts/operational/operational.png)
