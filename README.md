iTunes podcast RSS feed extractor
=================================

Every now and then I run into a link to a podcast in the Apple store,
and from there it's pretty hard to figure out the corresponding RSS
feed.

For example,
[https://podcasts.apple.com/us/podcast/estado-de-la-naci%C3%B3n-informe-2019/id1486121503](this)
is a typical webpage for such a podcast. If you dig around you'll find
links to the media, but not a link to the corresponding RSS feed that
you can use in your favorite podcast player.

There's a way to obtain that feed:

* Figure out the podcast ID
* Use the ID to query the podcast information in iTunes
* Extract the RSS feed from the result

This application automates that.

Usage
-----

Simply pass the Apple podcasts link as an argument:

```
$ go get github.com/mem/itunes-to-rss

$ itunes-to-rss https://podcasts.apple.com/us/podcast/estado-de-la-naci%C3%B3n-informe-2019/id1486121503
Estado de la Nación - Informe 2019 https://feeds.transistor.fm/estado-de-la-nacion-informe-2019
```

If you don't pass any arguments, a web server will be started on port
3000.
