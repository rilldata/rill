import { ANY_TYPES, NUMERICS } from "@rilldata/web-local/lib/duckdb-data-types";
import type { ExprCall } from "pgsql-ast-parser";
import type { ProfileColumn } from "@rilldata/web-local/lib/types";

// Taken from https://duckdb.org/docs/sql/aggregates
export const AllowedAggregates: {
  [agg in string]: Array<Set<string>>;
} = {
  // number, any
  arg_max: [NUMERICS, ANY_TYPES],
  argMax: [NUMERICS, ANY_TYPES],
  max_by: [NUMERICS, ANY_TYPES],
  arg_min: [NUMERICS, ANY_TYPES],
  argMin: [NUMERICS, ANY_TYPES],
  min_by: [NUMERICS, ANY_TYPES],

  // number
  avg: [NUMERICS],
  favg: [NUMERICS],
  fsum: [NUMERICS],
  sumKahan: [NUMERICS],
  kahan_sum: [NUMERICS],
  sum: [NUMERICS],
  max: [NUMERICS],
  min: [NUMERICS],
  kurtosis: [NUMERICS],
  mad: [NUMERICS],
  median: [NUMERICS],
  mode: [NUMERICS],
  skewness: [NUMERICS],
  stddev_pop: [NUMERICS],
  stddev_samp: [NUMERICS],
  var_pop: [NUMERICS],
  var_samp: [NUMERICS],

  // TODO: these need more detailed validation.
  // number, float
  quantile_cont: [NUMERICS, undefined],
  quantile_disc: [NUMERICS, undefined],
  approx_quantile: [NUMERICS, undefined],
  // number, number, number
  reservoir_quantile: [NUMERICS, undefined, undefined],

  // number, number
  covar_pop: [NUMERICS, NUMERICS],
  corr: [NUMERICS, NUMERICS],
  regr_avgx: [NUMERICS, NUMERICS],
  regr_avgy: [NUMERICS, NUMERICS],
  regr_count: [NUMERICS, NUMERICS],
  regr_intercept: [NUMERICS, NUMERICS],
  regr_r2: [NUMERICS, NUMERICS],
  regr_slope: [NUMERICS, NUMERICS],
  regr_sxx: [NUMERICS, NUMERICS],
  regr_sxy: [NUMERICS, NUMERICS],
  regr_syy: [NUMERICS, NUMERICS],

  // any
  count: [ANY_TYPES],
  approx_count_distinct: [ANY_TYPES],
  entropy: [ANY_TYPES],
};

export interface InvalidAggregate {
  name: string;
  aggregateNotAllowed?: boolean;
  invalidArgs?: Array<string>;
}

// TODO: validate arguments
export function validateAggregate(
  expr: ExprCall,
  args: Array<string>,
  profileColumns: Array<ProfileColumn>
): InvalidAggregate {
  if (!(expr.function.name in AllowedAggregates)) {
    return {
      name: expr.function.name,
      aggregateNotAllowed: true,
    };
  }

  return validateAggregateArguments(expr.function.name, args, profileColumns);
}

function validateAggregateArguments(
  name: string,
  args: Array<string>,
  profileColumns: Array<ProfileColumn>
): InvalidAggregate {
  const argTypes = args.map(
    (arg) => profileColumns.find((column) => column.name === arg)?.type
  );
  const valid =
    args.length !== AllowedAggregates[name].length ||
    AllowedAggregates[name].every(
      (argTypesSet, index) =>
        argTypes[index] === "" ||
        argTypesSet === undefined ||
        argTypes[index] === undefined ||
        argTypesSet.has(argTypes[index])
    );

  if (!valid) {
    return {
      name,
      invalidArgs: argTypes.map((argType) => (!argType ? "UNKNOWN" : argType)),
    };
  }
  return undefined;
}
