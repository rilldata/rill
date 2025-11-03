---
title: "Set up MCP and more!"
sidebar_label: "MCP Setup and Customization"
hide_table_of_contents: false
sidebar_position: 20
tags:
    - Quickstart
    - Tutorial
---

Connecting your Rill project to your AI Agent has never been easier! 

## Prerequisites

Before you begin, make sure you have:

- An MCP Client (Claude Desktop, ChatGPT Desktop, etc.)
- **Access to a [Rill Project](https://ui.rilldata.com/)** 

For detailed steps, please refer to our documentation on [MCP Servers](/explore/mcp)

## Step 1: Access your Rill Project's AI tab.

<img src='/img/explore/mcp/project-ai.png' class='rounded-gif'/>
<br />

In your project's AI tab, you'll be able to create a user token (personal access token) and copy the MCP config.json directly in the UI. Simply paste this into your application of choice.

:::tip Learn more about user tokens
For details on creating, managing, and using personal access tokens, see [User Tokens](/manage/user-tokens).
:::



## Step 2: Add config.json into Client 
Depending on your client application, paste the config from step 1.

- [Anthropic's Claude Desktop Documentation](https://modelcontextprotocol.io/quickstart/user)
- [OpenAI's ChatGPT Documentation](https://platform.openai.com/docs/guides/tools-remote-mcp#page-top)
- [Cursor Documentation](https://docs.cursor.com/context/model-context-protocol)
- [Windsurf Documentation](https://docs.windsurf.com/windsurf/cascade/mcp)

### Working with OpenAI's ChatGPT

To use Rill with OpenAI's ChatGPT, you'll need to follow the same steps as above to set up your MCP connection. Once that's done, you can start querying your Rill metrics directly from ChatGPT interfaces.

- Import your remote MCP servers directly in [ChatGPT settings](https://chatgpt.com/#settings) or in the desktop application navigate to settings by clicking on your profile picture in the bottom left corner.
- Connect your server in the Connectors tab. You may have to add the server as a source with the create button.

> Note: You may need at least the `Plus` plan to access connectors.

## Step 3: Start Querying your Agent about your Rill metrics.

A below is an example chat with Claude using our MCP server to get information on the commit history for one of our demo projects.

<img src='/img/explore/mcp/mcp-main.gif' class='rounded-gif'/>
<br />



## Step 4: Agent Customization
If you watched the whole clip above, you'll see that Claude tries to code a web app to visualize the information it found, but you'll notice it takes time (fast-forwarded in the clip). Instead, you can provide the Agent with context and instructions on how you'd like the Agent to behave. 

I highly recommend taking a look at our series of videos on [Conversation BI with Claude and Rill](https://www.youtube.com/playlist?list=PL_ZoDsg2yFKjSeetRNHbdI4GzmVn-XbBT).

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
