import React from 'react';
import PropTypes from 'prop-types';
import { useHistory } from '@docusaurus/router';

/**
 * ConnectorIcon component for documentation Connectors.
 * Displays an icon, header, content, and multiple action links.
 * Main card links to demo, with additional links for reference.
 */
function ConnectorIcon({ icon, header, content, link, linkLabel, target, rel, referenceLink }) {
    const history = useHistory();

    const handleReferenceClick = (e) => {
        e.stopPropagation();
        e.preventDefault();

        const anchorId = referenceLink.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '');
        const url = `/reference/project-files/connectors#${anchorId}`;

        // Use Docusaurus router to navigate
        history.push(url);

        // Try to scroll to the element after navigation
        setTimeout(() => {
            const element = document.getElementById(anchorId);
            if (element) {
                element.scrollIntoView({ behavior: 'smooth' });
            } else {
                // Try alternative anchor formats
                const alternatives = [
                    anchorId,
                    anchorId.replace(/-/g, ''),
                    referenceLink.toLowerCase(),
                    referenceLink.toLowerCase().replace(/\s+/g, '')
                ];

                for (const alt of alternatives) {
                    const el = document.getElementById(alt);
                    if (el) {
                        el.scrollIntoView({ behavior: 'smooth' });
                        break;
                    }
                }
            }
        }, 300);
    };

    return (
        <a className="Connector-icon" href={link} target={target} rel={rel}>
            {icon && (
                <div className="Connector-icon-icon">
                    {icon}
                </div>
            )}
            <div className="Connector-icon-content">
                {content}
            </div>
            <div className="Connector-icon-footer">
                <div className="Connector-icon-actions">
                    {referenceLink && (
                        <a
                            href={`/reference/project-files/connectors#${referenceLink.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '')}`}
                            className="Connector-icon-action-link"
                            onClick={handleReferenceClick}
                        >
                            YAML Reference
                        </a>
                    )}
                </div>
                <span className="Connector-icon-link">{linkLabel}</span>
            </div>
        </a>
    );
}

ConnectorIcon.propTypes = {
    icon: PropTypes.node,
    header: PropTypes.string.isRequired,
    content: PropTypes.string.isRequired,
    link: PropTypes.string.isRequired,
    linkLabel: PropTypes.string,
    target: PropTypes.string,
    rel: PropTypes.string,
    referenceLink: PropTypes.string,
};

export default ConnectorIcon; 