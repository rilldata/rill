/**
 * Helper type derives a type with at least the specified keys from T.
 *
 * For example, if we have `type T = {foo: string, bar: number, baz: boolean}`,
 * then `AtLeast<T, "foo">` will match objects with at least the "foo" key, but
 * the other keys will be optional
 */
export type AtLeast<T, K extends keyof T> = Partial<T> & Pick<T, K>;

/**
 * Helper type based on the internal `Expand` type from the TypeScript.
 *
 * If types on actions and selectors ever look nasty,
 * it's probably because we're missing an `Expand` somewhere.
 *
 *
 * see https://stackoverflow.com/questions/57683303/how-can-i-see-the-full-expanded-contract-of-a-typescript-type
 */
export type Expand<T> = T extends infer O ? { [K in keyof O]: O[K] } : never;
