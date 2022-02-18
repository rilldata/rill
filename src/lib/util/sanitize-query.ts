export function sanitizeQuery(query:string, toLower = true) {
    // remove comments;
    let noComments = query
        .replace(/--.*\n/g, ' ');
    // remove double+ spaces, \ns.
    let output = noComments
        .replace(/\n/g, ' ')
        .replace(/\s\s+/g, ' ')
        .replace(/;/g, '').trim();
    if (toLower) {
        output = output.toLowerCase();
    }
    return output;
    
}