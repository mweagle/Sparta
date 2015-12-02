hugo-bootstrap
==============
A theme with bootstrap, bootswatch(optional), font-awesome, highlightjs

*NOTE: This theme is copied from Hyde-Y. Not everything is ported to bootstrap.
Feel free to make changes and open pull requests.*


<small>Forked from [Hyde-Y](https://github.com/enten/hyde-y)</small>

You can find a live site using this theme [here](http://mmrath.com/).

## Screenshot

![preview](https://github.com/mmrath/hugo-bootstrap/blob/master/images/screenshot.png)

## Installation

```
$ cd your_site_repo/
$ mkdir themes
$ cd themes
$ git clone https://github.com/mmrath/hugo-bootstrap
```

See the [official Hugo themes documentation](http://gohugo.io/themes/installing) for more info.

## Usage

This theme expects a relatively standard Hugo blog/personal site layout:
```
.
└── content
    ├── post
    |   ├── post1.md
    |   └── post2.md
    ├── code
    |   ├── project1.md
    |   ├── project2.md
    ├── license.md        // this is used in the sidebar footer link
    └── other_page.md
```

Just run `hugo --theme=hugo-bootstrap` to generate your site!

## Configuration

### Hugo

An example of what your site's `config.toml` could look like. All theme-specific parameters are under `[params]` and standard Hugo parameters are used where possible.

``` toml
# hostname (and path) to the root eg. http://spf13.com/
baseurl = "http://www.example.com"

# Site title
title = "sitename"

# Copyright
copyright = "(c) 2015 yourname."

# Language
languageCode = "en-EN"

# Metadata format
# "yaml", "toml", "json"
metaDataFormat = "yaml"

# Theme to use (located in /themes/THEMENAME/)
theme = "hugo-bootstrap"

# Pagination
paginate = 10
paginatePath = "page"

# Enable Disqus integration
disqusShortname = "your_disqus_shortname"

[permalinks]
    post = "/:year/:month/:day/:slug/"
    code = "/:slug/"

[taxonomies]
    tag = "tags"
    topic = "topics"

[author]
    name = "yourname"
    email = "yourname@example.com"

#
# All parameters below here are optional and can be mixed and matched.
#
[params]
    # You can use markdown here.
    brand = "foobar"
    topline = "few words about your site"
    footline = "code with <i class='fa fa-heart'></i>"

    # Text for the top menu link, which goes the root URL for the site.
    # Default (if omitted) is "Home".
    home = "home"

    # Select a syntax highight.
    # Check the static/css/highlight directory for options.
    highlight = "default"

    # Google Analytics.
    googleAnalytics = "Your Google Analytics tracking code"

    # Sidebar social links.
    github = "enten/hugo-boilerplate" # Your Github profile ID
    bitbucket = "" # Your Bitbucket profile ID
    linkedin = "" # Your LinkedIn profile ID (from public URL)
    googleplus = "" # Your Google+ profile ID
    facebook = "" # Your Facebook profile ID
    twitter = "" # Your Twitter profile ID
    youtube = ""  # Your Youtube channel ID
    flattr = ""  # populate with your flattr uid

[blackfriday]
    angledQuotes = true
    fractions = false
    hrefTargetBlank = false
    latexDashes = true
    plainIdAnchors = true
    extensions = []
    extensionmask = []

```

### Menu

Create `data/Menu.toml` to configure the sidebar navigation links. Example below.

```toml
[about]
    Name = "About"
    IconClass = "fa-info-circle"
    URL = "/about"

[posts]
    Name = "Posts"
    Title = "Show list of posts"
    URL = "/post"

[tags]
    Name = "Tags"
    Title = "Show list of tags"
    URL = "/tags"
```

### Foot menu

Create `data/FootMenu.toml` to configure the footer navigation links. Example below.

```toml
[license]
    Name = "license"
    URL = "/license"
```

## Tips

* If you've added `theme = "hugo-bootstrap"` to your `config.toml`, you don't need to keep using the `--theme=hugo-bootstrap` flag!
* Although all of the syntax highlight CSS files under the theme's `static/css/highlight` are bundled with the site, only the one you choose will be included in the page and delivered to the browser.
* Change the favicon by providing your own as `static/favicon.png` (and `static/touch-icon-144-precomposed.png` for Apple devices) in your site directory.
* Hugo makes it easy to override theme layout and behaviour, read about it [here](http://gohugo.io/themes/customizing).
* Pagination is set to 10 items by default, change it by updating `paginate = 10` in your `config.toml`.

## Changes and enhancements from the original theme

* Modified to work with bootstrap and bootswatch
* ...many other small layout tweaks!

## Attribution

Obviously largely a port of the awesome [Hyde-Y](https://github.com/enten/hyde-y) theme.

## Questions, ideas, bugs, pull requests?

All feedback is welcome! Head over to the [issue tracker](https://github.com/mmrath/hugo-bootstrap/issues).

## License

Open sourced under the [MIT license](https://github.com/enten/hyde-y/blob/master/LICENSE).
