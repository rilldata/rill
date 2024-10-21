import React from 'react';

export default function LoomVideo({ loomId }) {
  // No autoplay, video will start manually by user interaction
  const url = `https://www.loom.com/embed/${loomId}?autoplay=0&mute=0&hide_owner=true&hide_share=true&hide_title=true`;

  return (
    <div
      style={{
        position: 'relative',
        paddingBottom: '56.25%', // 16:9 aspect ratio
        height: 0,
        overflow: 'hidden',
        maxWidth: '100%',
        background: '#000',
      }}
    >
      <iframe credentialless="true"
        src={url}
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
        }}
        frameBorder="0"
        allow="fullscreen; picture-in-picture"
        allowFullScreen
      ></iframe>
    </div>
  );
}
