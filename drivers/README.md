## Drivers

Content Scraper can retrieve data from various types of sources, e.g. Reddit, Twitter, Metafilter, Tumblr, etc.
Each of these sources has a Driver library to perform the following:

- Query the source for data ("scrape")
- Persist the retrieved posts to the database, including its contents, created date, a "score", etc.
- Filter the posts in some user-defined way. (e.g. only publish posts that are in the top 20% of scores)
- Publish the posts into an RSS feed.

The main application will call the Driver, which must implement the drivers.IDriver interface, to perform
these functions.

## Driver components

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
- SourceConfig: Describes a stream(s) of source data that we're interested in retrieving. For
  example, a list of subreddits.
- Harvester: Controls the Scraper to retrieve posts, and stores them to the database.
- Persistence: A database layer to store/retrieve posts.
- Feed: Describes a "view" onto the stored information. E.g. A "funny images" Feed might include such
  information as a list of subreddits like "/r/funny" and "/r/gifs", and parameters describing on how
  to filter them so that only the top x% are viewed.
- Filter: Applies the Feed configuration criteria to decide whether posts are eligible for publication.
  E.g. it might mark posts that are "> 5 hours old" & "minimum score 200" & "scored higher than the the
  80th percentile of scores for this subreddit".
  It also handled de-duplication and other filtering.
- Renderer: Converts a post (chosen for publication) into an RSS feed element.

## Shared components

The drivers package offers some utility classes that each Driver may find useful. E.g.
- Library to record the last time that an (arbitrary) source was pulled from.
