import React from 'react';
import PropTypes from 'prop-types';

/**
 * TileIcon component for documentation tiles.
 * Displays a header, content, and multiple action links.
 * Main card links to demo, with additional links for GitHub and walkthrough.
 */
function TileIcon({ header, content, link, linkLabel = 'Learn more', target, rel, githubLink, walkthroughLink }) {
    return (
        <a className="tile-icon" href={link} target={target} rel={rel}>
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
                </div>
            </div>
        </a>
    );
}

TileIcon.propTypes = {
    header: PropTypes.string.isRequired,
    content: PropTypes.string.isRequired,
    link: PropTypes.string.isRequired,
    linkLabel: PropTypes.string,
    target: PropTypes.string,
    rel: PropTypes.string,
    githubLink: PropTypes.string,
    walkthroughLink: PropTypes.string,
};

export default TileIcon; 