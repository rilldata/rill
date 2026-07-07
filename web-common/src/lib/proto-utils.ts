import type { Value } from "node_modules/@bufbuild/protobuf/dist/esm/google/protobuf/struct_pb";

/** A protobuf-es oneof field as generated for a `case` selector. */
type OneofSelector =
  | { case: string; value: unknown }
  | { case: undefined; value?: undefined };
/** Keys of `M` whose value is a protobuf-es oneof selector. */
type OneofKeys<M> = {
  [K in keyof M]: NonNullable<M[K]> extends OneofSelector ? K : never;
}[keyof M];
/** The set of case names available on the oneof selector `O`. */
type OneofCase<O> = O extends { case: infer C extends string } ? C : never;
/** The value type carried by the `C` case of the oneof selector `O`. */
type OneofValueForCase<O, C> = O extends { case: C; value: infer V }
  ? V
  : never;

/**
 * Returns the value of a protobuf-es oneof field for the requested case, or
 * undefined if the oneof is unset or set to a different case. `oneofKey` selects
 * the oneof field (e.g. "case" on a CategoricalSummary, "resource" on a
 * Resource) and the return type is narrowed to the value type of that case.
 */
export function getOneofValue<
  M extends object,
  K extends OneofKeys<M>,
  C extends OneofCase<NonNullable<M[K]>>,
>(
  message: M | undefined,
  oneofKey: K,
  caseName: C,
): OneofValueForCase<NonNullable<M[K]>, C> | undefined {
  const oneof = message?.[oneofKey] as OneofSelector | undefined;
  if (oneof?.case !== caseName) {
    return undefined;
  }
  return oneof.value as OneofValueForCase<NonNullable<M[K]>, C>;
}

export function valueAsNumber(val: Value | undefined) {
  return getOneofValue(val, "kind", "numberValue") ?? 0;
}
