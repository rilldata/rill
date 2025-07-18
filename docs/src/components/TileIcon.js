import React from 'react';
import PropTypes from 'prop-types';

/**
 * TileIcon component for documentation tiles.
 * Displays an icon, header, content, and multiple action links.
 * Main card links to demo, with additional links for GitHub, walkthrough, and reference.
 */
function TileIcon({ icon, header, content, link, linkLabel = 'Learn more', target, rel, githubLink, walkthroughLink, referenceLink }) {
    return (
        <a className="tile-icon" href={link} target={target} rel={rel}>
            {icon && (
                <div className="tile-icon-icon">
                    {icon}
                </div>
            )}
            <div className="tile-icon-header">{header}</div>
            <div className="tile-icon-content">
                {content}
            </div>
            <div className="tile-icon-footer">
                <span className="tile-icon-link">{linkLabel}</span>
                <div className="tile-icon-actions">
                    {githubLink && (
                        <a
                            href={githubLink}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="tile-icon-action-link"
                            onClick={(e) => e.stopPropagation()}
                        >
                            GitHub
                        </a>
                    )}
                    {walkthroughLink && (
                        <a
                            href={walkthroughLink}
                            className="tile-icon-action-link"
                            onClick={(e) => e.stopPropagation()}
                        >
                            Walkthrough
                        </a>
                    )}
                    {referenceLink && (
                        <a
                            href={referenceLink}
                            className="tile-icon-action-link"
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

TileIcon.propTypes = {
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

export default TileIcon; 