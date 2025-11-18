import { useColorMode } from '@docusaurus/theme-common';
import Navbar from '@theme-original/Navbar';
import { useEffect, useLayoutEffect } from 'react';

const MOBILE_ICON_LINKS = [
  {
    href: 'https://github.com/rilldata/rill',
    label: 'GitHub',
    src: '/icons/Github.svg',
  },
  {
    href: '/notes',
    label: 'Release Notes',
    src: '/icons/ReleaseNotes.svg',
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
  img.width = 24;
  img.height = 24;

  anchor.appendChild(img);
  return anchor;
}

function ensureSidebarIcons() {
  const brand = document.querySelector('.navbar-sidebar__brand');
  // Check if icons already exist
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
  // We only need colorMode to determine which icon to show.
  // The toggle logic is handled by the original button's onClick.
  const { colorMode } = useColorMode();

  // Handle Dark Mode Toggle Icons
  // useLayoutEffect fires synchronously before paint, reducing flicker
  useLayoutEffect(() => {
    const toggleButtons = document.querySelectorAll('button[class*="toggleButton"]');
    if (!toggleButtons.length) return;

    toggleButtons.forEach((btn, index) => {
      // Add ID to first toggle for potential targeting
      if (index === 0) btn.id = 'dark-mode-toggle';

      // Find or create the container for our custom icon
      let iconContainer = btn.querySelector('.icon-container');
      if (!iconContainer) {
        iconContainer = document.createElement('span');
        iconContainer.className = 'icon-container';
        
        // Clear existing Docusaurus toggle content (text/emojis)
        btn.innerHTML = ''; 
        btn.appendChild(iconContainer);
      }

      // Update Icon based on CURRENT mode (show the OPPOSITE icon)
      // If Dark Mode -> Show Sun (to switch to Light)
      // If Light Mode -> Show Moon (to switch to Dark)
      const isDark = colorMode === 'dark';
      
      iconContainer.innerHTML = `
        <img 
          src="/icons/${isDark ? 'Sun' : 'Moon'}.svg" 
          alt="${isDark ? 'Switch to light mode' : 'Switch to dark mode'}" 
          width="24" 
          height="24" 
        />
      `;
      
      btn.setAttribute('aria-label', isDark ? 'Switch to light mode' : 'Switch to dark mode');
    });
  }, [colorMode]);

  // Handle Mobile Sidebar Icons
  useEffect(() => {
    // Observer to inject icons when the mobile menu opens/renders
    const observer = new MutationObserver(() => {
      ensureSidebarIcons();
    });

    observer.observe(document.body, { childList: true, subtree: true });
    ensureSidebarIcons(); // Initial check

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
