export const NEW_MODEL_FILE_CONTENT = `
select a, b, c
from table
`;

export const NEW_DASHBOARD_FILE_CONTENT = `
kind: dashboard

title: Dashboard Title
table: table_name
timeseries: timestamp_column
default_time_range: P7D
dimensions:
  - name: dimension_name
    label: Dimension Name
    column: column_name
    description: Description
measures:
  - label: Measure Label
    expression: count(*)
`;

export const NEW_API_FILE_CONTENT = `
kind: api

sql:
  select a, b, c
  from table
`;

export const NEW_CHART_FILE_CONTENT = `
kind: chart

...
`;

export const NEW_THEME_FILE_CONTENT = `
kind: theme
colors:
  primary: crimson 
  secondary: lime 
`;

export const NEW_REPORT_FILE_CONTENT = `
kind: report

...
`;

export const NEW_ALERT_FILE_CONTENT = `
kind: alert

...
`;
