#!/usr/bin/python3

"""
    expand_templates.py

    MediaWiki API Demos
    Demo of `Expandtemplates` module: Expand the Project:Sandbox template.

    MIT License
"""

import sys, requests

lang = "sv"
template = "{{ordningstal|{{Stat/Sverige/Kommuner/Befolkning rank|0184}}}}"

#lang = sys.argv[1]
#template = sys.argv[2]


S = requests.Session()

URL = "https://%s.wikipedia.org/w/api.php" % lang

PARAMS = {
    "action": "expandtemplates",
    #"text": "{{Project:Sandbox}}",
    "text": template,
    "prop": "wikitext",
    "format": "json"
}

R = S.get(url=URL, params=PARAMS)
DATA = R.json()

print(DATA['expandtemplates']['wikitext'])
