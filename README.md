# dboxpaper

![dboxpaper](https://raw.githubusercontent.com/mattn/dboxpaper/master/dboxpaper-logo256.png)

client for Dropbox Paper

## Usage

```
NAME:
   dboxpaper - Dropbox Paper client

USAGE:
   dboxpaper [global options] command [command options] [arguments...]
   
VERSION:
   0.0.1
   
AUTHOR(S):
   mattn <mattn.jp@gmail.com> 
   
COMMANDS:
     cat         Cat paper
     delete      Delete paper permanently
     list, ls    Show papers
     upload, up  Upload paper
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

list papers

```
$ dboxpaper list
```

cat paper

```
$ dboxpaper cat XXXXXXXXXXX
```

upload paper

```
$ cat README.md | dboxpaper upload
```

update paper

```
$ cat README.md | dboxpaper upload XXXXXXXXXXX
```

delete paper

```
$ dboxpaper delete XXXXXXXXXXX
```

## Installation

```
$ go get github.com/mattn/dboxpaper
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
