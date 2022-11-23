export const metricsTemplate = `
display_name: "Sample Dashboard"
description: "a description that appears in the UI"

# model
#optional to declare this, otherwise it is the model.sql file in the same directory
from: ""

# populate with the first datetime type in the OBT
timeseries: ""

# default to opionated option around estimated timegrain,
# first in order is default time grain
timegrains:
  - "DAY"
# the timegrain that users will see when they first visit the dashboard.
default_timegrain:
  - "DAY"

# measures
# measures are presented in the order that they are written in this file.
measures: []

# dimensions
# dimensions are presented in the order that they are written in this file.
dimensions: []
`;
