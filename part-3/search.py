#!/usr/bin/env python3
"""
Search engine for FastMCP documentation using minsearch
"""

import zipfile
from pathlib import Path
from minsearch import Index


def extract_md_files(zip_path="fastmcp-main.zip"):
    """
    Extract and process .md and .mdx files from the FastMCP zip archive.

    Returns:
        list: List of documents with 'content' and 'filename' fields
    """
    documents = []

    with zipfile.ZipFile(zip_path, 'r') as zip_ref:
        # Get all file names in the zip
        file_list = zip_ref.namelist()

        # Filter for .md and .mdx files
        md_files = [f for f in file_list if f.endswith('.md') or f.endswith('.mdx')]

        print(f"Found {len(md_files)} markdown files")

        for file_path in md_files:
            # Read the content
            with zip_ref.open(file_path) as f:
                content = f.read().decode('utf-8')

            # Remove the first part of the path (fastmcp-main/)
            # Split by '/' and rejoin without the first part
            path_parts = file_path.split('/')
            if len(path_parts) > 1:
                cleaned_filename = '/'.join(path_parts[1:])
            else:
                cleaned_filename = file_path

            # Skip if it's empty after removing prefix
            if not cleaned_filename:
                continue

            documents.append({
                'content': content,
                'filename': cleaned_filename
            })

    print(f"Processed {len(documents)} documents")
    return documents


def create_search_index(documents):
    """
    Create and fit a minsearch index with the documents.

    Args:
        documents: List of documents with 'content' and 'filename' fields

    Returns:
        Index: Fitted minsearch index
    """
    # Create the index with 'content' as text field and 'filename' as keyword field
    index = Index(
        text_fields=['content', 'filename'],
        keyword_fields=[]
    )

    # Fit the index with documents
    index.fit(documents)

    print(f"Index created with {len(documents)} documents")
    return index


def search(query, index, num_results=5):
    """
    Search for documents using the query.

    Args:
        query: Search query string
        index: Fitted minsearch index
        num_results: Number of results to return (default: 5)

    Returns:
        list: Top num_results most relevant documents
    """
    # Boost content more than filename for relevance
    # boost_dict = {
    #     'content': 1.0,
    #     'filename': 2.0  # Filename matches are more important
    # }

    results = index.search(
        query=query,
        # boost_dict=boost_dict,
        num_results=num_results
    )

    return results


def main():
    """Main function to test the search implementation"""

    print("=" * 60)
    print("FastMCP Documentation Search Engine")
    print("=" * 60)
    print()

    # Step 1: Extract documents
    print("Step 1: Extracting markdown files from zip...")
    documents = extract_md_files()
    print()

    # Step 2: Create index
    print("Step 2: Creating search index...")
    index = create_search_index(documents)
    print()

    # Step 3: Test searches
    print("Step 3: Testing search functionality...")
    print()

    test_queries = [
        "demo",
    ]

    for query in test_queries:
        print(f"Query: '{query}'")
        print("-" * 60)

        results = search(query, index, num_results=5)

        if results:
            for i, doc in enumerate(results, 1):
                print(f"{i}. {doc['filename']}")
                # Show first 150 characters of content
                preview = doc['content'][:150].replace('\n', ' ')
                print(f"   Preview: {preview}...")
                print()
        else:
            print("No results found.")
            print()

        print()


if __name__ == "__main__":
    main()
