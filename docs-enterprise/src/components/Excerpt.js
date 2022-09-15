import React from "react";

export default function Excerpt({ text }) {
  return (
    <div
      style={{
        fontSize: 20,
        lineHeight: 1.5,
      }}
    >
      {text}
      <hr></hr>
    </div>
  );
}
