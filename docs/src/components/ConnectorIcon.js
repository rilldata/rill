import React from 'react';
import PropTypes from 'prop-types';

/**
 * ConnectorIcon component for documentation Connectors.
 * Displays an icon, header, content, and multiple action links.
 * Main card links to demo, with additional links for GitHub, walkthrough, and reference.
 */
function ConnectorIcon({ icon, header, content, link, linkLabel = 'Learn more', target, rel, githubLink, walkthroughLink, referenceLink }) {
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
                <span className="Connector-icon-link">{linkLabel}</span>
                <div className="Connector-icon-actions">
                    {githubLink && (
                        <a
                            href={githubLink}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="Connector-icon-action-link"
                            onClick={(e) => e.stopPropagation()}
                        >
                            GitHub
                        </a>
                    )}
                    {walkthroughLink && (
                        <a
                            href={walkthroughLink}
                            className="Connector-icon-action-link"
                            onClick={(e) => e.stopPropagation()}
                        >
                            Walkthrough
                        </a>
                    )}
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
    githubLink: PropTypes.string,
    walkthroughLink: PropTypes.string,
    referenceLink: PropTypes.string,
};

export default ConnectorIcon; 