/**
 * The type definition for a "profile column"
 */
 interface ProfileColumn {
    name: string;
    type: string;
    summary?: any; // FIXME
}

export interface Item {
    id: string;
}

export interface Query extends Item {
    /**  */
    query: string;
    /** sanitizedQuery is always a 1:1 function of the query itself */
    sanitizedQuery: string;
    /** name is used for the filename and exported file */
    name: string;
    /** cardinality is the total number of rows of the previewed dataset */
    cardinality?: number;
    /** sizeInBytes is the total size of the previewed dataset. 
     * It is not generated until the user exports the query.
     * */
    sizeInBytes?: number; // TODO: make sure this is just size
    error?: string;
    sources?: string[];
    profile?: ProfileColumn[]; // TODO: create Profile interface
    preview?: any;
    destinationProfile?: any;
}

export interface Source extends Item {
    id: string;
    path: string;
    name: string;
    profile: ProfileColumn[]; // TODO: create Profile interface
    head: any[];
    cardinality?: number;
    sizeInBytes?: number;
    numericalSummaries?:any;
    timestampSummaries?:any;
    categoricalSummaries?:any;
    nullCounts?:any;
}

export interface DataModellerState {
    activeQuery?: string;
    queries: Query[];
    sources: Source[];
    status: string;
}