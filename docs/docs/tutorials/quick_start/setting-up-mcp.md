---
title: "Set up MCP and more!"
sidebar_label: "MCP Setup and Customization"
hide_table_of_contents: false
tags:
    - Quickstart
---

Connecting your Rill project to your AI Agent has never been easier! 

## Prerequisites

Before you begin, make sure you have:

- A MCP Client (Claude Desktop, ChatGPT Desktop, etc.)
- **Access to a [Rill Project](https://ui.rilldata.com/)** 

For detailed steps, please refer to our documentation on [MCP Servers](/explore/mcp)

## Step 1: Access your Rill Project's AI tab.

<img src='/img/explore/mcp/project-ai.png' class='rounded-gif'/>
<br />

In your project's AI tab, you'll be able to create a [user based token] and copy the MCP config.json directly in the UI. Simple paste this into your application of choice.



## Step 2: Add config.json into Client 
Depending on your client application, paste the config from step 1.

- [Anthropic's Claude Desktop Documentation](https://modelcontextprotocol.io/quickstart/user)
- [OpenAI's ChatGPT Documentation](https://platform.openai.com/docs/guides/tools-remote-mcp#page-top)
- [Cursor Documetnation](https://docs.cursor.com/context/model-context-protocol)
- [Windsurf Documentation](https://docs.windsurf.com/windsurf/cascade/mcp)



## Step 3: Start Querying your Agent about your Rill metrics.

A below is an example chat with Claude using our MCP server to get information on the commit history for one of our demo projects.

<img src='/img/explore/mcp/mcp-main.gif' class='rounded-gif'/>
<br />



## Step 4: Agent Customization
If you watched the whole clip above, you'll see that Claude tries to make code a webapp to visualize the information that it found, but you'll see it takes time. (fast-forwarded in the clip) Instead, what you can do is provide the Agent context and instructions on how you'd like the Agent to behave. 

I highly recommend taking a look at our series of videos on [Conversation BI with Claude and Rill](https://www.youtube.com/playlist?list=PL_ZoDsg2yFKjSeetRNHbdI4GzmVn-XbBT
).

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/3xBCOY6rnsM?si=uvhUe11-at9c5bUh"
    frameBorder="0"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
    allowFullScreen
    style={{
      position: "absolute",
      top: 0,
      left: 0,
      width: "100%",
      height: "100%",
      borderRadius: "10px", 
    }}
  ></iframe>
</div>
