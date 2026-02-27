import React from 'react';
import Admonition from '@theme/Admonition';

/**
 * EnvPullTip component for connector documentation.
 * Displays the standard "Did you know?" tip about pulling credentials from Rill Cloud.
 */
function EnvPullTip() {
    return (
        <Admonition type="tip" title="Did you know?">
            <p>
                If this project has already been deployed to Rill Cloud and credentials have been set for this connector,
                you can use <code>rill env pull</code> to{' '}
                <a href="/developers/build/connectors/credentials#rill-env-pull">pull these cloud credentials</a>{' '}
                locally (into your local <code>.env</code> file). Please note that this may override any credentials
                you have set locally for this source.
            </p>
        </Admonition>
    );
}

export default EnvPullTip;
