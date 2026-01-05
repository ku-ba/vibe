#!/usr/bin/env python3
"""Direct MCP client to test the server"""

import asyncio
import json
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

async def test_mcp_server():
    """Test the MCP server by calling download_webpage tool"""

    # Server parameters
    server_params = StdioServerParameters(
        command="uv",
        args=["--directory", "/home/kuba/vibe/part-3", "run", "main.py"],
    )

    async with stdio_client(server_params) as (read, write):
        async with ClientSession(read, write) as session:
            # Initialize the session
            await session.initialize()

            # List available tools
            tools = await session.list_tools()
            print("Available tools:")
            for tool in tools.tools:
                print(f"  - {tool.name}: {tool.description}")
            print()

            # Call the download_webpage tool
            url = "https://github.com/alexeygrigorev/minsearch"
            print(f"Calling download_webpage with URL: {url}")

            result = await session.call_tool(
                "download_webpage",
                arguments={"url": url}
            )

            # Print results
            print(f"\nResult type: {type(result)}")
            print(f"Content length: {len(result.content[0].text)} characters")
            print(f"\nFirst 500 characters:")
            print("=" * 60)
            print(result.content[0].text[:500])
            print("=" * 60)

if __name__ == "__main__":
    asyncio.run(test_mcp_server())
