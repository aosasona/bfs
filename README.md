# BFS

A little CLI experiment to see if I can write something custom that runs almost as fast as Grep with Go constructs. This is intended to be reused in another project aimed at making search suck less on Windows (for myself at least).

> While it is _fairly_ safe, there are various ways you can still blow it up, this is a crappy version, use at your own risk.

## Running

```bash
go run . -root /Users/johndoe -query <search_query> [-json]
```

It supports returning JSON in case you need to evoke the binary from within your own code and then parse the response as it is streamed in.
