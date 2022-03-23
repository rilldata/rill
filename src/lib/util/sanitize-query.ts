export function sanitizeQuery(query: string, toLower = true) {
    // remove comments;
    let noComments = query
        .replace(/--.*/g, ' ');
    // remove double+ spaces, \ns.
    let output = noComments
        .replace(/\n/g, ' ')
        .replace(/\s\s+/g, ' ')
        .replace(/,\s+/g, ',')
        .replace(/;/g, '').trim();
    if (toLower) {
        output = output.toLowerCase();
    }
    // disallow anything other than SELECT and CTEs.
    if (!(output.match(/^SELECT/i) || output.match(/^WITH/i))) {
        output = '';
    }
    return output;

}