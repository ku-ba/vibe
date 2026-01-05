import requests

# Test the Jina reader with the example URL
url = "https://datatalks.club"
jina_url = f"https://r.jina.ai/{url}"

print(f"Testing Jina reader with URL: {url}")
print(f"Fetching from: {jina_url}\n")

try:
    response = requests.get(jina_url, timeout=30)
    response.raise_for_status()

    content = response.text
    print(f"Success! Retrieved {len(content)} characters")
    print("\nFirst 500 characters of content:")
    print("=" * 50)
    print(content[:500])
    print("=" * 50)

except requests.exceptions.RequestException as e:
    print(f"Error: {str(e)}")
