#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
This script extracts licensing info.
"""

import base64
import json
import os
import re
import subprocess
import sys
import traceback
import urllib.parse
import urllib.request


def get(url, cache, headers=None):
    """
    Fetches the given URL (with reply caching & optional authorization)
    """
    if url in cache:
        return cache[url]
    result = '{}'
    for _ in range(10):
        if headers is None:
            request = urllib.request.Request(url)
        else:
            request = urllib.request.Request(url, headers=headers)
        try:
            reply = urllib.request.urlopen(request)
            if reply.getcode() == 200:
                result = reply.read().decode('utf-8')
                break
        except urllib.error.HTTPError:
            print(traceback.format_exc(), file=sys.stderr)
            print(f'Retrying for {url}', file=sys.stderr)
            continue
        except ssl.SSLEOFError:
            print(traceback.format_exc(), file=sys.stderr)
            print(f'Retrying for {url}', file=sys.stderr)
            continue
    cache[url] = result
    return result


def authorization_credentials():
    """
    Reads authorization credentials from environment variables
    """
    user = os.environ.get('GH_API_USER')
    token = os.environ.get('GH_API_TOKEN')
    print(len(user), file=sys.stderr)
    print(len(token), file=sys.stderr)
    return (user, token)


def authorization_headers(user, token):
    """
    Prepares authorization header for an API request
    """
    if None in [user, token]:
        return None
    credentials = '%s:%s' % (user, token)
    auth = base64.b64encode(credentials.encode('utf-8'))
    headers = {'Authorization': 'Basic %s' % auth.decode('utf-8')}
    return headers


def run(command, path):
    """
    Runs a command in a given directory and returns its output
    or its cached output
    """
    if os.path.isdir(path):
        process = subprocess.run(
            command,
            encoding='utf-8',
            stdout=subprocess.PIPE,
            cwd=path,
            check=True)
        return process.stdout
    with open(path, 'r') as handle:
        return handle.read()


def execute(suffix, directory):
    """
    Executes licensing extraction script
    """
    base = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    script = os.path.basename(sys.argv[0]).replace('.py', suffix)
    script = os.path.join(base, 'automation', script)
    command = f'python3 "{script}"'
    source = os.path.join(base, directory)
    os.system(f'{command} "{source}"')


def github_normalize(url):
    """
    Normalizes GitHub URL
    """
    if url is None:
        return None
    url = url.replace('https://www.', '')
    url = url.replace('.git', '')
    url = re.sub(r'git.*://', '', url)
    url = url.replace('git@', '')
    return f'https://{url}'


def github_account(normalized_url):
    """
    Extracts account identifier from a GitHub repository URL
    """
    if normalized_url is None:
        return None
    return normalized_url.split('/')[3]


def github_owner(account, cache, headers=None):
    """
    Fetches the owner name of a GitHub account
    """
    url = f'https://api.github.com/users/{account}'
    profile = json.loads(get(url, cache, headers=headers))
    profile_name = profile['name'] if 'name' in profile else account
    return profile_name


def github_license(license_name, cache, headers=None):
    """
    Fetches the license template via GitHub API
    """
    license_name = urllib.parse.quote(license_name)
    license_url = f'https://api.github.com/licenses/{license_name}'
    license_description = json.loads(get(license_url, cache, headers=headers))
    return license_description['body']


def github_repository_license(normalized_url, cache, headers=None):
    """
    Fetches the licensing information of a GitHub repository
    """
    url = normalized_url.replace(
        'github.com', 'api.github.com/repos') + '/license'
    return json.loads(get(url, cache, headers=headers))


def escape(text):
    """
    Normalizes whitespace and escapes HTML tags
    """
    replacements = {
        '<': '&lt;',
        '>': '&gt;',
        '\t': ' ',
        '\f': '',
        '\v': '',
        '\xA0': '',
        '\x85': ''}
    for key in replacements:
        text = text.replace(key, replacements[key])
    return text


if __name__ == '__main__':
    execute('_go.py', 'core')
    execute('_js.py', 'ui')
