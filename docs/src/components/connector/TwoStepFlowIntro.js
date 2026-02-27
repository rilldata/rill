import React from 'react';
import PropTypes from 'prop-types';

/**
 * TwoStepFlowIntro component for connector documentation.
 * Displays the standard two-step flow explanation for adding data via the UI.
 *
 * @param {string} connector - The display name of the connector (e.g., "Athena", "BigQuery")
 * @param {string} step1Description - Optional custom description for step 1 (default: "Set up your {connector} connector with credentials")
 * @param {string} step2Description - Optional custom description for step 2 (default: "Define which database, table, or query to execute")
 */
function TwoStepFlowIntro({ connector, step1Description, step2Description }) {
    const defaultStep1 = `Set up your ${connector} connector with credentials`;
    const defaultStep2 = 'Define which database, table, or query to execute';

    return (
        <div className="two-step-flow-intro">
            <p>When you add data from {connector} through the Rill UI, the process follows two steps:</p>
            <ol>
                <li><strong>Configure Authentication</strong> - {step1Description || defaultStep1}</li>
                <li><strong>Configure Data Model</strong> - {step2Description || defaultStep2}</li>
            </ol>
            <p>This two-step flow ensures your credentials are securely stored in the connector configuration, while your data model references remain clean and portable.</p>
        </div>
    );
}

TwoStepFlowIntro.propTypes = {
    connector: PropTypes.string.isRequired,
    step1Description: PropTypes.string,
    step2Description: PropTypes.string,
};

export default TwoStepFlowIntro;
