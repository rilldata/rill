import json
import os

import vl_convert as vlc
from fastmcp import FastMCP
from fastmcp.utilities.types import Image
from openai import OpenAI

openai_client = OpenAI(api_key=os.environ.get("OPENAI_API_KEY"))

viz_mcp = FastMCP(
    name="RillVizServer",
)


@viz_mcp.tool()
def generate_chart(data: dict, prompt: str) -> Image:
    """Calls OpenAI to generate a chart for the provided data and prompt."""

    # Convert the data to a JSON string for including in the prompt
    data_json = json.dumps(data, indent=2)

    # Create a message for OpenAI with instructions to generate a Vega-Lite chart
    messages = [
        {
            "role": "system",
            "content": "You are a data visualization expert. Generate a Vega-Lite chart specification based on the provided data and user prompt.",
        },
        {
            "role": "user",
            "content": f"Data: {data_json}\n\nPrompt: {prompt}\n\nGenerate a Vega-Lite specification (version 5) for a chart that addresses this prompt using the provided data. Return ONLY valid JSON for the Vega-Lite specification, nothing else.",
        },
    ]

    # Call OpenAI API to generate the chart specification
    response = openai_client.chat.completions.create(
        model="gpt-4o",  # Use an appropriate model
        messages=messages,
        response_format={"type": "json_object"},
    )

    # Extract the Vega-Lite specification from the response
    try:
        vega_spec = json.loads(response.choices[0].message.content)

        # Use vl-convert to render the spec to PNG
        img_bytes = vlc.vegalite_to_png(
            vl_spec=json.dumps(vega_spec),
            scale=1.0,
        )

        # Return using FastMCP's Image helper
        return Image(data=img_bytes, format="png")
    except Exception as e:
        # If there's an error parsing the OpenAI response, say so
        print(f"Error generating chart with OpenAI: {str(e)}")
        raise e


viz_mcp._mcp_server.instructions = "This server provides access to RillData Viz APIs. "

if __name__ == "__main__":
    viz_mcp.run()
