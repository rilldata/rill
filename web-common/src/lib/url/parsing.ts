const conditionRegex = /(\w+)\s+(eq|ne)\s+(.+)/;
const dimensionValueRegex = /[‘'’]([^‘'’]*)[‘'’]/g;

export function parseFilterString(filterString: string, dimensions: string[]) {
  const initDimensions = new Map<
    string,
    { exclude: boolean; values: string[] }
  >();

  let errorMessage: string | null = null;

  if (!filterString || !dimensions.length) {
    return {
      initDimensions,
      errorMessage,
    };
  }

  if (
    (filterString.startsWith(`"`) && filterString.endsWith(`"`)) ||
    (filterString.startsWith(`'`) && filterString.endsWith(`'`)) ||
    (filterString.startsWith(`“`) && filterString.endsWith(`”`))
  ) {
    filterString = filterString.slice(1, -1);
  }

  const conditions = filterString.split(" and ");

  console.log({ conditions });

  conditions.forEach((condition) => {
    console.log({ condition });
    const match = condition.match(conditionRegex);
    if (match) {
      const [, dimension, operator, valueString] = match;

      const values: string[] = [];

      if (valueString.startsWith("(") && valueString.endsWith(")")) {
        let match: RegExpExecArray | null;

        while ((match = dimensionValueRegex.exec(valueString)) !== null) {
          values.push(match[1]);
        }
      } else if (valueString.match(dimensionValueRegex)) {
        values.push(valueString.slice(1, -1));
      } else {
        errorMessage = `Value missing quotes: ${valueString}`;
        return;
      }

      if (!dimensions.includes(dimension)) {
        errorMessage = `Invalid dimension: ${dimension}`;
        return;
      } else if (values.length === 0) {
        errorMessage = `Invalid values: ${valueString}`;
        return;
      }

      const exclude = operator === "ne" || operator === "nin";

      initDimensions.set(dimension, {
        exclude,
        values,
      });
    } else {
      errorMessage = `Invalid condition. Expected format: <dimension> <eq|ne> ('<value>', '<value>')`;
    }
  });

  return {
    initDimensions,
    errorMessage,
  };
}
