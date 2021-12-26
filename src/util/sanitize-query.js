export function sanitizeQuery(query) {
    // remove comments;
    let noComments = query
        .replace(/--.*\n/g, ' ')
    // remove double+ spaces, \ns.
    let output = noComments
        .replace(/\n/g, ' ')
        .replace(/\s\s+/g, ' ')
        .replace(/;/g, '').trim().toLowerCase();
    return output;
    
}