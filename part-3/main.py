from fastmcp import FastMCP
import requests

mcp = FastMCP("Demo 🚀")

@mcp.tool
def add(a: int, b: int) -> int:
    """Add two numbers"""
    return a + b

@mcp.tool
def download_webpage(url: str) -> str:
    """
    Download the content of any web page in markdown format using Jina reader.

    Args:
        url: The URL of the web page to download

    Returns:
        The content of the web page in markdown format
    """
    # Prepend r.jina.ai to the URL
    jina_url = f"https://r.jina.ai/{url}"

    try:
        response = requests.get(jina_url, timeout=30)
        response.raise_for_status()
        return response.text
    except requests.exceptions.RequestException as e:
        return f"Error downloading webpage: {str(e)}"

if __name__ == "__main__":
    mcp.run()
