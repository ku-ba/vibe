#!/usr/bin/env python3
"""Direct test of the download_webpage function"""

import requests

def download_webpage(url: str) -> str:
    """Download webpage content using Jina reader"""
    jina_url = f"https://r.jina.ai/{url}"
    try:
        response = requests.get(jina_url, timeout=30)
        response.raise_for_status()
        return response.text
    except requests.exceptions.RequestException as e:
        return f"Error downloading webpage: {str(e)}"

# Test with the example URL
url = "https://datatalks.club"
print(f"Downloading content from: {url}\n")

content = download_webpage(url)

print(f"Retrieved {len(content)} characters\n")
print("First 1000 characters:")
print("=" * 60)
print(content[:1000])
print("=" * 60)
