---
title: Storeman
actions:
  - storeman
---

[Storeman](start-app://harbour-storeman) is a community-run, third‑party app store. You can also
browse it on the web at [openrepos.net](https://openrepos.net).

## Installation

1. Open [Storeman Installer](https://openrepos.net/content/olf/storeman-installer) in your browser
   (preferably the [Sailfish Browser](start-app://sailfish-browser), not an Android browser) and download
   the latest version.
   - You can save it anywhere, for example `Home folder -> Downloads`.
2. Open the downloaded file. You should see an install prompt—confirm it.
   - If you dismissed the prompt, open the file again using a [file browser](../basic/file-browsing.md).
3. Wait.

The last step is crucial and can take a couple of minutes. The file you downloaded is only an installer. It
checks your system and then installs the right Storeman app. If you still don’t see Storeman on your
home screen after about two minutes, try rebooting your phone.

## Security

OpenRepos and Storeman are **not as safe as the official store**. This does *not* mean that everything
there is dangerous—personally, I have not seen malware—but the system is based on trust. You should only
install apps from developers you trust, and you should never install things blindly.

### The trust model

> If you don’t want to read the wall of text below, the TL;DR is: Storeman apps can break your system
> (by accident or on purpose), so always read the description and comments and make sure you trust the
> developer.

Each developer has their own repository (think of it as their personal app shelf). By default, no third‑party
repositories are enabled, so you can’t install anything yet. When you find an app you want, you must first
use the pull‑down menu and tap `Add repository`. This makes *all apps from that developer* available on your
phone.

After that, you can install the app.

Important detail: Jolla Store and Storeman both use the same system‑wide installer. Storeman simply *adds*
extra repositories. This means an app from Storeman can replace an app from the Jolla Store if it has a
higher version number. In rare cases, that can **break your system**—usually by accident, but it is
possible.

That said, bad things don’t happen all the time. Most Storeman apps work fine, and most phones stay stable.
The risks above are uncommon but important to understand.

## Usage

If the **Security** section didn’t scare you off, welcome! Storeman is easy to use: browse apps, search,
read comments, and leave comments.

When you want to install something, remember the one extra step for new developers:
1. Open the pull‑down menu and tap `Add repository` (only needed the first time for that developer).
2. Open the pull‑down menu again and tap `Install`.
