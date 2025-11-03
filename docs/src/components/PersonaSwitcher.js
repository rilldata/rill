import React, { useState, useEffect } from 'react';
import '../css/_persona-switcher.scss';

export function usePersona() {
    const [persona, setPersonaState] = useState('developer');

    useEffect(() => {
        // Load persona from localStorage
        const savedPersona = localStorage.getItem('rill-docs-persona');
        if (savedPersona) {
            setPersonaState(savedPersona);
        }

        // Set initial data attribute on body
        document.body.setAttribute('data-persona', savedPersona || 'developer');
    }, []);

    const setPersona = (newPersona) => {
        localStorage.setItem('rill-docs-persona', newPersona);
        setPersonaState(newPersona);
        // Update body data attribute for debugging and potential CSS hooks
        document.body.setAttribute('data-persona', newPersona);
        // Dispatch custom event so other components can react
        window.dispatchEvent(new CustomEvent('persona-change', { detail: newPersona }));
    };

    return [persona, setPersona];
}

export default function PersonaSwitcher() {
    const [persona, setPersona] = usePersona();

    return (
        <div className="persona-switcher">
            <button
                className={`persona-btn ${persona === 'developer' ? 'active' : ''}`}
                onClick={() => setPersona('developer')}
                aria-label="Switch to Developer view"
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <polyline points="16 18 22 12 16 6"></polyline>
                    <polyline points="8 6 2 12 8 18"></polyline>
                </svg>
                <span>Developer</span>
            </button>
            <button
                className={`persona-btn ${persona === 'business' ? 'active' : ''}`}
                onClick={() => setPersona('business')}
                aria-label="Switch to Business User view"
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="3" y="3" width="7" height="7"></rect>
                    <rect x="14" y="3" width="7" height="7"></rect>
                    <rect x="14" y="14" width="7" height="7"></rect>
                    <rect x="3" y="14" width="7" height="7"></rect>
                </svg>
                <span>Business User</span>
            </button>
        </div>
    );
}

