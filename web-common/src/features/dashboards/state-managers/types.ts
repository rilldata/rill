/**
 * Helper type derives a type with at least the specified keys from T.
 *
 * For example, if we have `type T = {foo: string, bar: number, baz: boolean}`,
 * then `AtLeast<T, "foo">` will match objects with at least the "foo" key, but
 * the other keys will be optional
 */
export type AtLeast<T, K extends keyof T> = Partial<T> & Pick<T, K>;
