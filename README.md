# Beginner's Guide for SailfishOS backend

This backend transforms the .md files in the [docs](docs) directory into JSON api
that the [app](https://github.com/RikudouSage/SfosBeginnerGuide) knows how to handle.

## Guide

The [docs](docs) directory contains language tags as the only elements (for example, [en](docs/en)) which
in turn contains .md files and other subdirectories.

Each page needs some metadata, at the very least a `title`. See [the English intro](docs/en/index.md) for
an example.

Additionally, these metadata are supported:

- `links`: a simple array of strings with links to relevant content, the title will be fetched automatically
- `actions`: a simple array of action ids which will be available on the page inside the app
  - the support for every action must be developed inside the app
