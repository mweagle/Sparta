---
date: 2018-02-24 06:29:36
title: Meta
weight: 0
---

This outlines how to edit the documentation itself

## Requirements

  * The [Sparta repo](https://github.com/mweagle/Sparta)
  * Hugo v.0.36.1:

    ```
    $ hugo version
    Hugo Static Site Generator v0.36.1 darwin/amd64 BuildDate: 2018-02-15T09:07:45Z
    ```

Note that the required Hugo version will be automatically installed by the `docs` branch [Makefile](https://github.com/mweagle/Sparta/blob/docs/Makefile).

## Editing

  * `git checkout` the _docs_ branch
  * Start a preview server with `make edit`:
    ```
    $ hugo server --disableFastRender

                      | EN
    +------------------+-----+
      Pages            |  50
      Paginator pages  |   1
      Non-page files   |   0
      Static files     | 457
      Processed images |   0
      Aliases          |   3
      Sitemaps         |   1
      Cleaned          |   0

    Total in 774 ms
    Watching for changes in ./github.com/mweagle/Sparta/{content,layouts,static,themes}
    Serving pages from memory
    Web Server is available at //localhost:1313/ (bind address 127.0.0.1)
    Press Ctrl+C to stop
    ```

  * Visit http://localhost:1313
  * Edit the _/content_ subdirectory contents
  * Push your _/docs_ branch to GitHub and open a PR

Visit the [docdock](http://docdock.netlify.com/) site for complete documentation regarding shortcodes and included libraries.
