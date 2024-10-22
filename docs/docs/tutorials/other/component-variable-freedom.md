---
title: "Using Variables in Components, to what extent?"
sidebar_label: "Filter your Components for your own personal View"
sidebar_position: 10
hide_table_of_contents: false
tags:
  - Canvas Dashboard
  - Canvas Component
---

## An idea

What if you create a component that is COMPLETELY based off of variables? 
The source is selectable and depending on which source your selecting, this filters the filters on which columns are measures are available. 


selector_source.yaml -> outputs 'source_name' to all components

source_name -> selector_dimensions.yaml -> outputs dimensions -> all components 
source_name -> selector_measure.yaml -> outputs measures -> all components

source_name and dimension -> selector_group_by -> outputs a groupby for components defaults to ""

source_name, measure
x, measure_0
x, measure_1
y, measure_0
y, measure_1

