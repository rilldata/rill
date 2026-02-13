import React from 'react';
import PropTypes from 'prop-types';
import Admonition from '@theme/Admonition';

/**
 * DeployToCloud component for connector documentation.
 * Displays the standard "Deploy to Rill Cloud" section with connector-specific details.
 *
 * Note: The heading should be added in markdown (## Deploy to Rill Cloud)
 * so it appears in the ToC. This component only renders the body content.
 *
 * @param {string} connector - The display name of the connector (e.g., "Athena", "BigQuery")
 * @param {string} connectorId - The lowercase connector ID for the reference link (e.g., "athena", "bigquery")
 * @param {string} credentialDescription - Description of what credentials are needed (e.g., "an access key and secret for an AWS service account")
 */
function DeployToCloud({ connector, connectorId, credentialDescription }) {
    return (
        <>
            <p>
                When deploying a project to Rill Cloud, Rill requires you to explicitly provide {credentialDescription}{' '}
                with access to {connector} used in your project. Please refer to our{' '}
                <a href={`/reference/project-files/connectors#${connectorId}`}>connector YAML reference docs</a>{' '}
                for more information.
            </p>
            <p>
                If you subsequently add sources that require new credentials (or if you simply entered the wrong
                credentials during the initial deploy), you can update the credentials by pushing the <code>Deploy</code>{' '}
                button to update your project or by running the following command in the CLI:
            </p>
            <pre><code>rill env push</code></pre>

            <Admonition type="tip" title="Did you know?">
                <p>
                    If you've already configured credentials locally (in your <code>&lt;RILL_PROJECT_DIRECTORY&gt;/.env</code> file),
                    you can use <code>rill env push</code> to{' '}
                    <a href="/developers/build/connectors/credentials#rill-env-push">push these credentials</a>{' '}
                    to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials
                    automatically by running <code>rill env pull</code>.
                </p>
            </Admonition>
        </>
    );
}

DeployToCloud.propTypes = {
    connector: PropTypes.string.isRequired,
    connectorId: PropTypes.string.isRequired,
    credentialDescription: PropTypes.string.isRequired,
};

export default DeployToCloud;
