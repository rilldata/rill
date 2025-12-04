---
title: "Miscellaneous Widgets"
sidebar_label: "Miscellaneous"
sidebar_position: 20
---

import ImageCodeToggle from '@site/src/components/ImageCodeToggle';

Miscellaneous widgets in Rill Canvas provide additional functionality for text, images, and other non-data elements. These widgets help enhance your dashboards with rich content. For more information, refer to our [Components reference doc](/reference/project-files/component).

## Text/Markdown

Text widgets allow you to add formatted text, markdown content, and documentation directly to your dashboards.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/text.png"
  imageAlt="Text component showing markdown formatting"
  code={`- markdown:
      alignment:
        horizontal: left
        vertical: middle
      content: |-
        # Markdown
        *Italic*  
        **Bold**  
        ***Bold Italic***  
        ~~Strikethrough~~

        [Rill Home](https://rilldata.com)
        Inline code: \`print("Hello")\`

        Block code:
        \`\`\`python
        def greet():
            return "Hello, Markdown!"
      width: 6`}
  codeLanguage="yaml"
/>

## Image

Image widgets let you embed images, logos, and visual elements into your dashboards. Put files in your public/ folder and reference them directly in the `url` as `public/image.png`.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/image.png"
  imageAlt="Image component showing embedded logo"
  code={`- image:
      url: https://cdn.prod.website-files.com/659ddac460dbacbdc813b204/660b0f85094eb576187342cf_rill_logo_sq_gradient.svg
    width: 6`}
  codeLanguage="yaml"
/>

## Component

Reference a reusable component created outside of the canvas dashboard.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/component.png"
  imageAlt="Table showing detailed data columns"
  code={`- component: my_component`}
  codeLanguage="yaml"
/>