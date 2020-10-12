#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
This script extracts licensing info associated with js code.
"""

import json
import sys
import urllib.parse

import licenses


URL_AFL = 'https://web.archive.org/web/20061013013632if_/http://opensource.linux-mirror.org/licenses/afl-2.1.txt'  # noqa: E501, pylint: disable=C0301
URL_CC_ZERO_1 = 'https://creativecommons.org/publicdomain/zero/1.0/legalcode.txt'  # noqa: E501
URL_CC_BY_3 = 'https://creativecommons.org/licenses/by/3.0/legalcode.txt'  # noqa: E501
CC_PREFIX = 'Copyright (c) [year] [fullname]\nNo changes were made to this package.\n'  # noqa: E501


def versionless(package):
    """
    Removes the version from the package reference
    """
    return package[:1+package[1:].find('@')]


def own(package):
    """
    Check if the package associated with this repository
    """
    return package.startswith('phytofy-ui')


def license_present(packages, package):
    """
    Checks if the license file is present
    """
    if 'licenseFile' in packages[package]:
        path = packages[package]['licenseFile']
        if path.lower().find('readme.m') == -1:
            return True
    return False


def extract_js_repository(packages, package, cache):
    """
    Obtains the repository for the package
    """
    name = packages[package]['name']
    identifier = urllib.parse.quote(name).replace('/', '%2F')
    url = f'https://api.npms.io/v2/package/{identifier}'
    npm_package = licenses.get(
        url, cache, headers={'user-agent': 'curl/7.71.1'})
    npm_package = json.loads(npm_package)
    try:
        url = npm_package['collected']['metadata']['repository']['url']
    except KeyError:
        url = None
    packages[package]['url'] = licenses.github_normalize(url)


def extract_js_repository_account(packages, package):
    """
    Obtains the account of the owner of the package
    """
    packages[package]['account'] = licenses.github_account(
        packages[package]['url'])


def extract_js_repository_owner(packages, package, cache, headers=None):
    """
    Obtains the name of the owner of the package
    """
    name = packages[package]['name']
    account = packages[package]['account']
    if account is None:
        owner = None
    else:
        owner = licenses.github_owner(account, cache, headers)
    if owner is None:
        owner = f'author(s) of {name}'
    packages[package]['owner'] = owner


def extract_js_package_license(packages, package, cache, headers=None):
    """
    Formats licensing information from a template for a package without one
    """
    package_licenses = packages[package]['licenses']
    if not isinstance(package_licenses, list):
        package_licenses = [package_licenses]
    text = []
    for package_license in package_licenses:
        package_license = package_license.replace('*', '')
        if packages[package]['licenses'] == 'Public Domain':
            template = ''
        elif package_license == 'AFLv2.1':
            template = licenses.get(URL_AFL, cache)
        elif package_license == 'CC0-1.0':
            template = licenses.get(
                URL_CC_ZERO_1, cache, headers={'user-agent': 'curl/7.71.1'})
        elif package_license == 'CC-BY-3.0':
            template = licenses.get(
                URL_CC_BY_3, cache, headers={'user-agent': 'curl/7.71.1'})
            template = f'{CC_PREFIX}{template}'
        elif package_license == 'BSD':
            replacements = {
                'glob-to-regexp': 'BSD-2-Clause',
                'json-schema': 'BSD-3-Clause'}
            package_license = replacements[packages[package]['name']]
            template = licenses.github_license(package_license, cache, headers)
        else:
            template = licenses.github_license(package_license, cache, headers)
        owner = packages[package]['owner']
        text.append(template.replace(
            '[year]', '2020').replace('[fullname]', owner))
    packages[package]['text'] = '\n---\n'.join(text)


def extract_js_packages(packages):
    """
    Obtains licensing information for each of the packages
    """
    cache = {}
    credentials = licenses.authorization_credentials()
    headers = licenses.authorization_headers(*credentials)
    for package in packages:
        if license_present(packages, package):
            with open(packages[package]['licenseFile'], 'r') as handle:
                packages[package]['text'] = handle.read()
        else:
            packages[package]['name'] = versionless(package)
            extract_js_repository(packages, package, cache)
            extract_js_repository_account(packages, package)
            extract_js_repository_owner(packages, package, cache, headers)
            extract_js_package_license(packages, package, cache, headers)
        package_licenses = packages[package]['licenses']
        if isinstance(package_licenses, list):
            package_licenses = ', '.join(package_licenses)
        package_licenses = package_licenses.replace('(', '').replace(')', '')
        package_licenses = package_licenses.replace(' OR ', ' or ')
        package_licenses = package_licenses.replace(' AND ', ' and ')
        packages[package]['licenses'] = package_licenses


def format_licensing(packages):
    """
    Formats the licensing information for the given package
    """
    for package in packages:
        license_name = packages[package]['licenses']
        license_text = licenses.escape(packages[package]['text'])
        template = '<b>%s</b>; License - %s:<br/>\n<pre>\n%s\n</pre><br/><br/>'
        print(template % (package, license_name, license_text))


def process(packages):
    """
    Processes entries of the package listing
    """
    extract_js_packages(packages)
    format_licensing(packages)


if __name__ == '__main__':
    if len(sys.argv) != 2:
        print(f'Usage: {sys.argv[0]} JS_PATH')
        print('Example:')
        print(f'{sys.argv[0]} .')
        sys.exit(1)
    else:
        js_path = sys.argv[1]
        listing = licenses.run(['license-checker', '--json'], js_path)
        listing = json.loads(listing)
        own_packages = [package for package in listing.keys() if own(package)]
        for own_package in own_packages:
            del listing[own_package]
        process(listing)
