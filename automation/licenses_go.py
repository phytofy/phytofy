#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
This script extracts licensing info associated with go code.
"""

import re
import sys

import licenses


def format_licensing(name, url, version, cache):
    """
    Formats the licensing info for the given package
    """
    credentials = licenses.authorization_credentials()
    headers = licenses.authorization_headers(*credentials)
    license_github = licenses.github_repository_license(
        url, cache, headers=headers)
    license_name = license_github['license']['name']
    license_text = licenses.get(license_github['download_url'], cache)
    license_text = licenses.escape(license_text)
    template = '<b>%s %s</b>; License - %s:<br/>\n<pre>\n%s\n</pre><br/><br/>'
    print(template % (name, version, license_name, license_text))


def process(entries):
    """
    Processes entries of the package listing
    """
    cache = {}
    for url, version in entries:
        name = url
        if url.startswith('github.'):
            url = f'https://{url}'
            format_licensing(name, url, version, cache)
        if url.startswith('gopkg.'):
            url = f'https://{url}'
            result = licenses.get(url, cache).split('\n')
            line = [line for line in result if line.find(
                'Source Code') != -1][0]
            url = re.findall('href=[\"\'](.*?)[\"\']', line)[0]
            url = url[:url.find('/tree')]
            format_licensing(name, url, version, cache)


if __name__ == '__main__':
    if len(sys.argv) != 2:
        print(f'Usage: {sys.argv[0]} GO_PATH')
        print('Example:')
        print(f'{sys.argv[0]} .')
        sys.exit(1)
    else:
        go_path = sys.argv[1]
        listing = licenses.run(['go', 'list', '-m', 'all'], go_path)
        listing = listing.split('\n')[1:]
        listing = [tuple(item.split(' ')) for item in listing if item != '']
        process(listing)
