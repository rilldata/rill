interface DataItem {
  [key: string]: string | number | Date;
}

const isValidDate = (value: unknown) =>
  value instanceof Date && !isNaN(value.getTime());

function isTimeString(value: string): boolean {
  // Simple heuristic: check if the string matches an ISO date format
  return /^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}:\d{2}Z?)?$/.test(value);
}

export function identifyFields(data: DataItem[]): {
  timeFields: string[];
  nominalFields: string[];
  quantitativeFields: string[];
} {
  const timeFields = new Set<string>();
  const nominalFields = new Set<string>();
  const quantitativeFields = new Set<string>();

  // Determine the total number of unique keys across all data items
  const allKeys = new Set<string>();
  data.forEach((item) => {
    Object.keys(item).forEach((key) => allKeys.add(key));
  });

  // Early exit if there are no items or keys
  if (data.length === 0 || allKeys.size === 0) {
    return { timeFields: [], nominalFields: [], quantitativeFields: [] };
  }

  for (const item of data) {
    for (const key of allKeys) {
      const value = item[key];

      // Skip if the field is already classified
      if (
        timeFields.has(key) ||
        nominalFields.has(key) ||
        quantitativeFields.has(key)
      )
        continue;

      // Skip if value is null or undefined
      if (value === null || value === undefined) continue;

      if (isValidDate(value)) {
        timeFields.add(key);
      }
      if (typeof value === "string") {
        if (isTimeString(value)) {
          timeFields.add(key);
        } else {
          nominalFields.add(key);
        }
      } else if (typeof value === "number") {
        quantitativeFields.add(key);
      }
    }

    // Exit if all fields are classified before processing all items
    if (
      timeFields.size + nominalFields.size + quantitativeFields.size ===
      allKeys.size
    ) {
      break;
    }
  }

  return {
    timeFields: Array.from(timeFields),
    nominalFields: Array.from(nominalFields),
    quantitativeFields: Array.from(quantitativeFields),
  };
}

export function suggestChartTypes(
  timeFields: string[],
  nominalFields: string[],
  quantitativeFields: string[],
): string[] {
  const chartTypes: string[] = [];
  if (timeFields.length > 0) {
    chartTypes.push("line", "area", "bar");
  }
  if (nominalFields.length > 0) {
    chartTypes.push("stacked area", "grouped bar", "stacked bar");
  }
  return chartTypes;
}
