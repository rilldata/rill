import React from 'react';
import PropTypes from 'prop-types';
import { Highlight } from 'prism-react-renderer';
import { themes } from 'prism-react-renderer';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

/**
 * ImageCodeToggle component for documentation.
 * Displays a sample image and corresponding code side by side in two columns.
 */
function ImageCodeToggle({ image, code, imageAlt = "Sample image", codeLanguage = "yaml" }) {
    const { colorMode } = useDocusaurusContext();
    const theme = colorMode === 'dark' ? themes.dracula : themes.github;

    return (
        <div className="image-code-toggle">
            <div className="image-code-toggle-content">
                <div className="image-code-toggle-image">
                    <div className="image-code-toggle-header">
                        <svg height="15px" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path fillRule="evenodd" clipRule="evenodd" d="M2 3C2 2.44772 2.44772 2 3 2H10C10.5523 2 11 2.44772 11 3V12C11 12.5523 10.5523 13 10 13H3C2.44772 13 2 12.5523 2 12V3ZM4 4V11H9V4H4Z" fill="#008FD4"></path>
                            <path fillRule="evenodd" clipRule="evenodd" d="M13 3C13 2.44772 13.4477 2 14 2H21C21.5523 2 22 2.44772 22 3V8C22 8.55228 21.5523 9 21 9H14C13.4477 9 13 8.55228 13 8V3ZM15 4V7H20V4H15Z" fill="#008FD4"></path>
                            <path fillRule="evenodd" clipRule="evenodd" d="M13 12C13 11.4477 13.4477 11 14 11H21C21.5523 11 22 11.4477 22 12V21C22 21.5523 21.5523 22 21 22H14C13.4477 22 13 21.5523 13 21V12ZM15 13V20H20V13H15Z" fill="#008FD4"></path>
                            <path fillRule="evenodd" clipRule="evenodd" d="M2 16C2 15.4477 2.44772 15 3 15H10C10.5523 15 11 15.4477 11 16V21C11 21.5523 10.5523 22 10 22H3C2.44772 22 2 21.5523 2 21V16ZM4 17V20H9V17H4Z" fill="#008FD4"></path>
                        </svg>
                        <span>Preview</span>
                    </div>
                    <img src={image} alt={imageAlt} />
                </div>

                <div className="image-code-toggle-code">
                    <div className="image-code-toggle-header">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M9.4 16.6L4.8 12l4.6-4.6L8 6l-6 6 6 6 1.4-1.4zm5.2 0L19.2 12l-4.6-4.6L16 6l6 6-6 6-1.4-1.4z" />
                        </svg>
                        <span>Inline Code</span>
                    </div>
                    <Highlight
                        code={code}
                        language={codeLanguage}
                        theme={theme}
                    >
                        {({ className, style, tokens, getLineProps, getTokenProps }) => (
                            <pre className={className} style={style}>
                                {tokens.map((line, i) => (
                                    <div key={i} {...getLineProps({ line })}>
                                        {line.map((token, key) => (
                                            <span key={key} {...getTokenProps({ token })} />
                                        ))}
                                    </div>
                                ))}
                            </pre>
                        )}
                    </Highlight>
                </div>
            </div>
        </div>
    );
}

ImageCodeToggle.propTypes = {
    image: PropTypes.string.isRequired,
    code: PropTypes.string.isRequired,
    imageAlt: PropTypes.string,
    codeLanguage: PropTypes.string,
};

export default ImageCodeToggle;
