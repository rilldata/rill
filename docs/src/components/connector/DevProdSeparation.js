import React from 'react';

/**
 * DevProdSeparation component for connector documentation.
 * Displays the standard section about separating dev and prod environments.
 *
 * Note: The heading should be added in markdown (## Separating Dev and Prod Environments)
 * so it appears in the ToC. This component only renders the body content.
 */
function DevProdSeparation() {
    return (
        <>
            <p>
                When ingesting data locally, consider setting parameters in your connector file to limit how much
                data is retrieved, since costs can scale with the data source. This also helps other developers
                clone the project and iterate quickly by reducing ingestion time.
            </p>
            <p>
                For more details, see our <a href="/developers/build/connectors/templating">Dev/Prod setup docs</a>.
            </p>
        </>
    );
}

export default DevProdSeparation;
