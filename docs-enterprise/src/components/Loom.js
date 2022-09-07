import React from "react";

export default function Loom({ id, padding = "56.25%" }) {
  const url = "https://www.loom.com/embed/" + id;
  const pads = padding + " 0 0 0";
  return (
    <div
      style={{
        padding: pads,
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
