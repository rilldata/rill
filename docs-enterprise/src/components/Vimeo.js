import React from "react";

export default function Vimeo({ id }) {
  const url =
    "https://player.vimeo.com/video/" +
    id +
    "?muted=1&autopause=0&autoplay=1&loop=1&title=0&byline=0&portrait=0";
  return (
    <div
      style={{
        padding: "56.23% 0 0 0",
        position: "relative",
      }}
    >
      <iframe
        src={url}
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          width: "100%",
          height: "100%",
        }}
        frameBorder="0"
        autoPlay
        allowFullScreen
      ></iframe>
    </div>
  );
}
