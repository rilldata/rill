import { sanitizeQuery } from "./sanitize-query";

describe("sanitizeQuery", () => {
    it("removes comments, unused whitespace, and ;", () => {
        const output = sanitizeQuery(`
-- whatever this is
SELECT * from         whatever;
-- another extraneous comment.
`)
        expect(output).toBe('select * from whatever')
    })
    it("option to not lowercase a query", () => {
        const output = sanitizeQuery(`
-- whatever this is
SELECT * from         whateveR;
-- another extraneous comment.        
        `, false)
        expect(output).toBe('SELECT * from whateveR')
    })
    it("removes comments not ending in a newline", () => {
        const output = sanitizeQuery(`
SELECT * from         whateveR; -- inline comment`, false)
        expect(output).toBe('SELECT * from whateveR')
    })
})