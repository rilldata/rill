import { sanitizeQuery } from "./sanitize-query.mjs";

describe("sanitizeQuery", () => {
    it("removes comments, unused whitespace, and ;", () => {
        const output = sanitizeQuery(`
-- whatever this is
SELECT * from         whatever;
-- another extraneous comment.
`)
        expect(output).toBe('select * from whatever')
    })
})