# Architecture

(This app was partly an exercise in learning Go, so there may be some anti-patterns in play)

## Essential concepts

A *Content Source* is a website that provides some content, such as Reddit or Twitter. There are drivers
for each type of content source. Because each content source organizes its content differently, each
source type needs custom configuration. For example, the Reddit driver requires a list of subreddits,
and the Twitter drive requires a list of `@` accounts.

A *Feed* is a single stream of posts from a single content source type. Its contents
are source-specific; for example, a Feed from sub-reddits may actually glom together 1-many subreddits.
Users are expected to explore content by Feed. They are listed in the web server's main menu, and selecting
one will show posts from that Feed.

## Architecture overview
The program does the following:
* On startup it reads config from a YAML file. This is primarily a list of content sources, and any
  credentials required to retrieve content from those sources.
* It periodically runs a harvester, which pulls content from the each content source
  specified in the config file, and writes them to a local Sqlite database.
* It runs a local web server with a simple UI. This has two page types:
** A menu of Feeds (consolidated from all content sources)
** A Feed view (with optional page number). Each page lists ~10 items from that Feed.
   The displayed items are filtered & ranked according to a heuristic that tries
   to promote higher quality & newer content. The parameters for these can be tweaked in the config
   file.

## Config file

The config file parser code is in [config](/config).

The program needs a storage directory for the SQLite database. This location is hard-coded as
`$HOME/.contentscraper` (Linux) and `$LOCALAPPDATA/ContentScraper` (Windows).

This directory is also searched for the config file. This must be either located in the storage directory
and named `contentscraper.yaml` or `contentscraper.yml`, or an alternative path to the file can be
provided in the `-config` option.

## Database layer

The backend database is provided by Sqlite. This is accessed through the golang "standard" `database/sql`
interface, which supports generic operations for many different RDBMS backends.

The [database](/database) codebase does very little; it exists to support initializing new connections
and create in-memory databases to assist testing other components.

## Drivers

Each of the content sources has a Driver library to perform the following:

- Query the source for data ("scrape")
- Persist the retrieved posts to the database, including its contents, created date, a "score", etc.
- Filter the posts in some user-defined way. (e.g. only publish posts that are in the top 20% of scores)
- View the contents of a feed.

The main application will call a content source's Driver to perform these actions. The driver must
must implement the drivers.IDriver interface in [drivers/types.go](/drivers/types.go).

Each driver is initialized in a driver-specific way from the monolithic Config datastructure. It is
also typically handed a database connection so that it can retrieve/persist the harvested data.

### Driver components

Ideally the sources would share enough commonality that they can reuse code. However they are different
enough that there's very little to share (and Go's limited inheritance model doesn't help much either).
There is no set structure for a driver to adhere to, but each typically has the following component classes:

- Driver: An object that implements the IDriver interface. This concrete instance is typically a Facade
  that delegates to subclasses.
- Constructor: function(s) responsible for creating the Driver, including any dependent objects. This
  is called by the "main loop" on initialization. It's expected to use the (passed) Config object to
  configure how the Driver works.
- Scraper: A client library for retrieving data over the network from the source, and parsing it into
  a set of "posts". (e.g. graw library for Reddit, Twitter API, libCurl+HTML parser for Metafilter).
  There is typically a wrapper around this Scraper to marshall data into a more app-specific format
  and to handle cases like navigating "pagination".
- Harvester: Controls the Scraper to retrieve posts, and stores them to the database.
- Persistence: A database layer to store/retrieve posts.
- Filter: Applies the Feed configuration criteria to decide whether posts are eligible for publication.
  E.g. it might mark posts that are "> 5 hours old" & "minimum score 200" & "scored higher than the the
  80th percentile of scores for this subreddit". It also handled de-duplication and other filtering.
- Renderer: Converts a post (chosen for publication) into an RSS feed element.

### Harvesting

The harvester uses drivers to gather content from each type of source.  Then in the main harvesting loop
(which is kicked off periodically), each driver's `Harvest()` method is called in turn.
   
The harvesting is handled simply by a `Harvest()` function, which is periodically called from the
main loop. Each driver is responsible for retrieving the content from the configured sources, and
placing it into the database. Because each content source is different, a driver will typically use
its own database tables.

#### Scrapers

A Scraper is a component of the harvesting process that is responsible for retrieving
the content from the content source. They are content-source aware, meaning that they understand
how to:
* log into the content source
* understand pagination
* understand parsing the data into a suitable format

Scrapers tend to heavily rely on 3rd party libraries.

They typically pass back the content to the harvesting implementation, which then decides how to
persist the information and whether to continue scraping.

### Web server

Each driver is also responsible for serving the content. At startup time, each configured driver
is registered with the main web server. The driver specifies a base URL and a web handler method
that will display the contents of a Feed. The main web server registers the handler method under
the base URL, and the driver is responsible for content generation after that.

The following driver methods are used by the main web server:
* `GetBaseUrlPath()` uniquely identifies a driver and is also used as a URL prefix for any Feeds
  handled by that driver.
* `GetHttpHandler()` tells the web server what method to call when the user visits a Feed page
  associated with a driver. The method is responsible for extracting the feed name and page number,
  and returning a rendered page.
* `GetFeeds()` returns a list of Feeds that this driver is responsible for. This is used in the UI
  to display a menu.
