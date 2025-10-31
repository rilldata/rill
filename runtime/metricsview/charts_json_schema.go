package metricsview

// ChartsJSONSchema defines the JSON schema for chart specifications,
// Used by the MCP create_chart tool to validate chart configurations.
const ChartsJSONSchema = `{
  "type": "object",
  "required": ["chart_type", "spec"],
  "properties": {
    "chart_type": {
      "type": "string",
      "description": "The type of chart to render."
    },
    "spec": {
      "$ref": "#/$defs/ChartSpec",
      "description": "The chart specification containing configuration and data references."
    }
  },
  "$defs": {
    "ChartSpec": {
      "type": "object",
      "required": ["metrics_view", "time_range"],
      "properties": {
        "metrics_view": {
          "type": "string",
          "description": "The metrics view to query data from."
        },
        "time_range": {
          "type": "object",
          "required": ["start", "end"],
          "properties": {
            "start": {
              "type": "string",
              "description": "Start time for the time range."
            },
            "end": {
              "type": "string",
              "description": "End time for the time range."
            }
          }
        },
        "where": {
          "$ref": "#/$defs/Expression",
          "description": "Optional expression for filtering the underlying data before aggregation."
        },
        "time_grain": {
          "$ref": "#/$defs/TimeGrain",
          "description": "Time grain for temporal aggregation."
        },
        "x": {
          "$ref": "#/$defs/FieldConfig",
          "description": "X-axis field configuration."
        },
        "y": {
          "$ref": "#/$defs/FieldConfig",
          "description": "Y-axis field configuration."
        },
        "y1": {
          "$ref": "#/$defs/FieldConfig",
          "description": "Y-axis field configuration."
        },
        "y2": {
          "$ref": "#/$defs/FieldConfig",
          "description": "Tertiary Y-axis field configuration."
        },
        "color": {
          "anyOf": [
            { "$ref": "#/$defs/FieldConfig" },
            { "type": "string" }
          ],
          "description": "Color field configuration or static color value."
        },
        "measure": {
          "$ref": "#/$defs/FieldConfig",
          "description": "Measure field configuration."
        },
        "stage": {
          "$ref": "#/$defs/FieldConfig",
          "description": "Stage field configuration for funnel charts."
        },
        "innerRadius": {
          "type": "number",
          "description": "Inner radius for donut charts."
        },
        "breakdownMode": {
          "type": "string",
          "description": "Breakdown mode for the chart."
        },
        "mode": {
          "type": "string",
          "description": "Display mode for the chart."
        },
        "show_data_labels": {
          "type": "boolean",
          "description": "Whether to show data labels on the chart."
        }
      }
    },
    "FieldConfig": {
      "type": "object",
      "required": ["field", "type"],
      "properties": {
        "field": {
          "type": "string",
          "description": "The field name from the metrics view."
        },
        "type": {
          "type": "string",
          "description": "The field type (dimension or measure)."
        },
        "fields": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "description": "Array of field names for multi-field configurations."
        },
        "colorMapping": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "value": {
                "type": "string",
                "description": "The value to map."
              },
              "color": {
                "type": "string",
                "description": "The color to use for this value."
              }
            }
          },
          "description": "Mapping of values to colors."
        },
        "colorRange": {
          "type": "object",
          "description": "Color range configuration for continuous scales."
        },
        "labelAngle": {
          "type": "number",
          "description": "Angle for axis labels in degrees."
        },
        "legendOrientation": {
          "type": "string",
          "description": "Orientation of the legend."
        },
        "limit": {
          "type": "number",
          "description": "Maximum number of values to display."
        },
        "max": {
          "type": "number",
          "description": "Maximum value for the axis."
        },
        "min": {
          "type": "number",
          "description": "Minimum value for the axis."
        },
        "mark": {
          "type": "string",
          "description": "Mark type for the field."
        },
        "showAxisTitle": {
          "type": "boolean",
          "description": "Whether to show the axis title."
        },
        "showNull": {
          "type": "boolean",
          "description": "Whether to show null values."
        },
        "showTotal": {
          "type": "boolean",
          "description": "Whether to show total values."
        },
        "sort": {
          "anyOf": [
            {
              "enum": [
                "x",
                "y",
                "-x",
                "-y",
                "color",
                "-color",
                "measure",
                "-measure"
              ],
              "type": "string"
            },
            {
              "items": {
                "type": "string"
              },
              "type": "array"
            }
          ],
          "description": "Sort order for the field values."
        },
        "timeUnit": {
          "type": "string",
          "description": "Time unit for temporal fields."
        },
        "zeroBasedOrigin": {
          "type": "boolean",
          "description": "Whether to use zero-based origin for the axis."
        }
      }
    },
    "Expression": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Expression name"
        },
        "val": {
          "description": "Expression value"
        },
        "cond": {
          "$ref": "#/$defs/Condition"
        },
        "subquery": {
          "$ref": "#/$defs/Subquery"
        }
      }
    },
    "Condition": {
      "type": "object",
      "properties": {
        "op": {
          "$ref": "#/$defs/Operator",
          "description": "Operator for the condition"
        },
        "exprs": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/Expression"
          },
          "description": "Expressions in the condition"
        }
      },
      "required": ["op"]
    },
    "Subquery": {
      "type": "object",
      "properties": {
        "dimension": {
          "$ref": "#/$defs/Dimension"
        },
        "measures": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/Measure"
          }
        },
        "where": {
          "$ref": "#/$defs/Expression"
        },
        "having": {
          "$ref": "#/$defs/Expression"
        }
      },
      "required": ["dimension", "measures"]
    },
    "Dimension": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Name of the dimension"
        },
        "compute": {
          "$ref": "#/$defs/DimensionCompute",
          "description": "Optionally configure a derived dimension, such as a time floor."
        }
      },
      "required": ["name"]
    },
    "DimensionCompute": {
      "type": "object",
      "properties": {
        "time_floor": {
          "$ref": "#/$defs/DimensionComputeTimeFloor"
        }
      }
    },
    "DimensionComputeTimeFloor": {
      "type": "object",
      "properties": {
        "dimension": {
          "type": "string",
          "description": "Dimension to apply time floor to"
        },
        "grain": {
          "$ref": "#/$defs/TimeGrain",
          "description": "Time grain for flooring"
        }
      },
      "required": ["dimension", "grain"]
    },
    "Measure": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "description": "Name of the measure"
        },
        "compute": {
          "$ref": "#/$defs/MeasureCompute",
          "description": "Optionally configure a derived measure, such as a comparison."
        }
      },
      "required": ["name"]
    },
    "MeasureCompute": {
      "type": "object",
      "properties": {
        "count": {
          "type": "boolean",
          "description": "Whether to compute count"
        },
        "count_distinct": {
          "$ref": "#/$defs/MeasureComputeCountDistinct"
        },
        "comparison_value": {
          "$ref": "#/$defs/MeasureComputeComparisonValue"
        },
        "comparison_delta": {
          "$ref": "#/$defs/MeasureComputeComparisonDelta"
        },
        "comparison_ratio": {
          "$ref": "#/$defs/MeasureComputeComparisonRatio"
        },
        "percent_of_total": {
          "$ref": "#/$defs/MeasureComputePercentOfTotal"
        },
        "uri": {
          "$ref": "#/$defs/MeasureComputeURI"
        }
      }
    },
    "MeasureComputeCountDistinct": {
      "type": "object",
      "properties": {
        "dimension": {
          "type": "string",
          "description": "Dimension to count distinct values for"
        }
      },
      "required": ["dimension"]
    },
    "MeasureComputeComparisonValue": {
      "type": "object",
      "properties": {
        "measure": {
          "type": "string",
          "description": "Measure to compare"
        }
      },
      "required": ["measure"]
    },
    "MeasureComputeComparisonDelta": {
      "type": "object",
      "properties": {
        "measure": {
          "type": "string",
          "description": "Measure to compute delta for"
        }
      },
      "required": ["measure"]
    },
    "MeasureComputeComparisonRatio": {
      "type": "object",
      "properties": {
        "measure": {
          "type": "string",
          "description": "Measure to compute ratio for"
        }
      },
      "required": ["measure"]
    },
    "MeasureComputePercentOfTotal": {
      "type": "object",
      "properties": {
        "measure": {
          "type": "string",
          "description": "Measure to compute percentage for"
        },
        "total": {
          "type": "number",
          "description": "Total value to use for percentage calculation"
        }
      },
      "required": ["measure"]
    },
    "MeasureComputeURI": {
      "type": "object",
      "properties": {
        "dimension": {
          "type": "string",
          "description": "Dimension to generate URI for"
        }
      },
      "required": ["dimension"]
    },
    "Operator": {
      "type": "string",
      "enum": [
        "",
        "eq",
        "neq",
        "lt",
        "lte",
        "gt",
        "gte",
        "in",
        "nin",
        "ilike",
        "nilike",
        "or",
        "and"
      ],
      "description": "Comparison or logical operator"
    },
    "TimeGrain": {
      "type": "string",
      "enum": [
        "TIME_GRAIN_UNSPECIFIED",
        "TIME_GRAIN_MILLISECOND",
        "TIME_GRAIN_SECOND",
        "TIME_GRAIN_MINUTE",
        "TIME_GRAIN_HOUR",
        "TIME_GRAIN_DAY",
        "TIME_GRAIN_WEEK",
        "TIME_GRAIN_MONTH",
        "TIME_GRAIN_QUARTER",
        "TIME_GRAIN_YEAR"
      ],
      "description": "Time granularity"
    }
  }
}`
