---
title: Storeman
actions:
  - storeman
---

[Storeman](start-app://harbour-storeman) is a third-party app store provided by the community. You can also
view it on the web at [openrepos.net](https://openrepos.net).

## Installation

1. Visit [Storeman Installer](https://openrepos.net/content/olf/storeman-installer) in your browser (preferably the [Sailfish Browser](start-app://sailfish-browser) 
   and not an Android one) and download the latest version. At the time of writing this guide it's
   [2.3.0-10](https://openrepos.net/sites/default/files/packages/5928/harbour-storeman-installer-2.3.0-release10.noarch.rpm).
   - You can choose any directory when downloading, for example `Home folder -> Downloads`
2. Open the downloaded file, and it should ask you whether you want to install it, do so
   - If you accidentally dismiss it, you need to open it using a [file browser](../basic/file-browsing.md)
3. Wait a while

The last step is crucial, it really is gonna take a while. What you downloaded was an installer for the actual
Storeman app, that installer will do some checks of what environment you're running in and then install the
appropriate Storeman. If after about 2 minutes you still see no Storeman on your homescreen
you can try rebooting your phone.

## Security

OpenRepos and Storeman are inherently **very insecure**. It doesn't that there are viruses or whatnot
(in fact I've never encountered a malicious app there), but the way it works means you should trust
the person you're installing apps from and shouldn't install anything blindly.

### The trust model

> If you don't want to read the wall of text below, the TL;DR is: Storeman apps can break your system,
> be it intentionally or by accident, you should always read the description and comments and make sure
> you trust the developer.

Every developer has their own repository where they publish the apps to. By default, there are no third-party
repositories enabled, and thus you cannot install anything yet. When you visit an app you like, you must first
use the pull-down menu to `Add repository` which will then make *all apps from the developer* available to you.

Then you can install them.

Note that at its core, all apps are installed the same way, be it Jolla Store or Storeman - Storeman just
makes the apps available to the system-wide installer when you add a repository. What this means is that
the apps can actually conflict between Jolla Store and Storeman and usually the one with the bigger version
number wins. That could potentially **break your system**, be it intentionally (which hasn't happened yet)
or accidentally (which *has* happened in the past).

That out of the way, it's not like bad things happen all the time when you install an app from Storeman,
apps from there work just fine, and systems don't usually break. The above threats are (for now?) mostly
theoretical but crucial to understand nonetheless. 

## Usage

If the **Security** section didn't scare you, welcome! The usage is dead simple – browse apps, search for apps,
read comments, write comments...

And if some app catches your eye, install them! If you read the Security section, you already know that before
installing an app you must add the app's repository from the pull-down menu. What this means is that the first
time you install an app from a new developer, you have to do one extra step before installing: Use pull-down
menu and choose `Add repository`.

After the repository is added, you use the pull-down menu and choose `Install`.
