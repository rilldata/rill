import React, { useEffect } from 'react';
import Navbar from '@theme-original/Navbar';
import { useColorMode } from '@docusaurus/theme-common';

const MOBILE_ICON_LINKS = [
  {
    href: 'https://github.com/rilldata/rill',
    label: 'GitHub',
    src: '/icons/Github.svg',
  },
  {
    href: 'https://www.rilldata.com/blog',
    label: 'Blog',
    src: '/icons/MessageSquareQuote.svg',
  },
];

function createIconLink({ href, label, src }) {
  const anchor = document.createElement('a');
  anchor.href = href;
  anchor.target = '_blank';
  anchor.rel = 'noopener noreferrer';
  anchor.className = 'mobile-nav-icon-link';
  anchor.setAttribute('aria-label', label);

  const img = document.createElement('img');
  img.src = src;
  img.alt = label;
  img.width = 20;
  img.height = 20;

  anchor.appendChild(img);
  return anchor;
}

function ensureSidebarIcons() {
  const brand = document.querySelector('.navbar-sidebar__brand');
  if (!brand || brand.querySelector('.mobile-nav-icon-links')) {
    return;
  }

  const container = document.createElement('div');
  container.className = 'mobile-nav-icon-links';

  MOBILE_ICON_LINKS.forEach((link) => {
    container.appendChild(createIconLink(link));
  });

  const closeButton = brand.querySelector('.navbar-sidebar__close');
  if (closeButton) {
    brand.insertBefore(container, closeButton);
  } else {
    brand.appendChild(container);
  }
}

export default function NavbarWrapper(props) {
  const { colorMode, setColorMode } = useColorMode();

  useEffect(() => {
    const toggleButton = document.getElementById('dark-mode-toggle');
    if (!toggleButton) {
      return undefined;
    }

    let iconContainer = toggleButton.querySelector('.icon-container');
    if (!iconContainer) {
      iconContainer = document.createElement('span');
      iconContainer.className = 'icon-container';
      toggleButton.innerHTML = '';
      toggleButton.appendChild(iconContainer);
    }

    const updateIcon = () => {
      if (colorMode === 'dark') {
        iconContainer.innerHTML = `
          <img src="/icons/Sun.svg" alt="Light mode" width="24" height="24" />
        `;
        toggleButton.setAttribute('aria-label', 'Switch to light mode');
      } else {
        iconContainer.innerHTML = `
          <img src="/icons/Moon.svg" alt="Dark mode" width="24" height="24" />
        `;
        toggleButton.setAttribute('aria-label', 'Switch to dark mode');
      }
    };

    updateIcon();

    const handleClick = () => {
      setColorMode(colorMode === 'dark' ? 'light' : 'dark');
    };

    toggleButton.addEventListener('click', handleClick);

    return () => {
      toggleButton.removeEventListener('click', handleClick);
    };
  }, [colorMode, setColorMode]);

  useEffect(() => {
    const observer = new MutationObserver(() => {
      ensureSidebarIcons();
    });

    observer.observe(document.body, { childList: true, subtree: true });
    ensureSidebarIcons();

    return () => {
      observer.disconnect();
      const existing = document.querySelector('.mobile-nav-icon-links');
      if (existing) {
        existing.remove();
      }
    };
  }, []);

  return <Navbar {...props} />;
}
