/**
 * swim-lane-placement.ts
 * ----------------------
 * This function will place a set of leaderboards (or any other list of DOM elements)
 * in such a way that 
 * (1) they are in distinct, rougly equally-sized columns;
 * (2) the columns can be flexbox, but won't reflow to the next;
 * 
 * This will look similar to a masonry layout, with a crucial difference; it will not reflow
 * across columns unless you explicitly ask it to.
 * This makes the Explorer load experience less jarring when a user clicks on something
 * and the leaderboard jumps to somewhere. I mean, where did it go?
 * The problem with the masonry approach is, users develop their own transient mental map
 * of where the leader board was. Changing the column is not good.
 */

function sum(total, a) { return total + a };

/**
 * 
 * @param elements an array of just about anything
 * @param sizeFunction a functino that will extract a number
 * @param columns the number of columns you expect to be using
 * @returns an array of arrays, which should contain fairly equally-weighted columns
 */
export function swimLanePlacement(elements, sizeFunction = element => element.getBoundingClientRect().height, columns = 3) {
        // first, extract the heights.
        const heights = elements.map(sizeFunction);
        const totalHeight = heights.reduce(sum, 0);
    
        const columnSet = Array.from({length: columns}).map(() => []);
    
        heights.forEach((height:number, i) => {
            const candidateHeights = heights.slice(0, i).reduce(sum, 0);
            const whichColumn = (~~(columns * candidateHeights / totalHeight));
            columnSet[whichColumn].push(elements[i]);
        });
        return columnSet;
}