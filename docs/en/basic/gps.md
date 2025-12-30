---
title: GPS
---

Getting the GPS to work (especially the first time) might not be really straightforward.
There are multiple ways to make it easier.

> If you're wondering why this is necessary while it's not on Android or iOS, it's because of the
> mass amount of data the corporations behind the OS have on everything and everyone – they use it
> to assist the GPS module with getting a fix faster.

## Fine-tune GPS settings

Go to `Settings -> Location` and switch the Accuracy to `Custom settings`. Afterwards tap
`Select custom settings` and enable the following options:

- GPS positioning
- Offline position lock

## Install offline data

You can either install the official `Positioning` packages from [Jolla Store](../installing-apps/jolla-store.md)
or the MLS Manager from [Storeman](../installing-apps/storeman.md).

Both are currently outdated, but the MLS Manager has more up-to-date data, so if you're fine with Storeman,
I recommend using that.

> The first time you're trying to get a GPS fix you should do so outside and give it a few minutes. That's
> even after you do the below. Subsequent location locking should be way faster.

### Official Positioning packages

1. Open the [Jolla Store](start-app://store-client)
2. In the pull-down menu select Search
3. Search for "Positioning" (without the quotes)
4. Install the package for your region

### MLS Manager

1. Open [Storeman](start-app://harbour-storeman)
2. In the pull-down menu select Search
3. Search for "MLS Manager"
4. Install it as per the [instructions](../installing-apps/storeman.md)
5. Launch it
6. Download packages for each country you're often in
    1. Installing is done by long holding on the country and selecting `Install`

## Troubleshooting

If you still can't get a fix, you might want to give the app [GPSInfo](start-app://harbour-gpsinfo) a try.
If you don't have it installed, simply search for it in the [Jolla Store](store-app://harbour-gpsinfo).
