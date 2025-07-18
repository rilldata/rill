import React from 'react';
import PropTypes from 'prop-types';

/**
 * ConnectorIcon component for documentation Connectors.
 * Displays an icon, header, content, and multiple action links.
 * Main card links to demo, with additional links for reference.
 */
function ConnectorIcon({ icon, header, content, link, linkLabel = 'Learn more', target, rel, referenceLink }) {
    return (
        <a className="Connector-icon" href={link} target={target} rel={rel}>
            {icon && (
                <div className="Connector-icon-icon">
                    {icon}
                </div>
            )}
            <div className="Connector-icon-header">{header}</div>
            <div className="Connector-icon-content">
                {content}
            </div>
            <div className="Connector-icon-footer">
                <div className="Connector-icon-actions">
                    {referenceLink && (
                        <a
                            href={referenceLink}
                            className="Connector-icon-action-link"
                            onClick={(e) => e.stopPropagation()}
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