Jump [ARCHIVED]
===============

This project is archived. Use a tool like fasd.

---

[![License: MIT](http://img.shields.io/badge/license-MIT-red.svg?style=flat-square)](http://opensource.org/licenses/MIT)

Jump enables you to jump to bookmarked locations in your filesystem from your
shell.

## Installation
Make sure you have [Go](http://golang.org) installed and a `GOPATH` configured.
You can then install (and update) Jump with:

```sh
go get -u github.com/cassava/jump
```

Because a program cannot change your shell working directory, a shell function
needs to do that. In your shell configuration, such as `~/.bashrc` or `~/.zshrc`
insert the following line:

```sh
source <(jump --source)
````

If you did not add `$GOPATH/bin` to your `PATH`, then use:

```sh
source <($GOPATH/bin/jump --source)
```

By default, this defines a function:

```sh
function jp() {
    eval $(jump $@)
}
```

Jump returns all output to stderr except for `cd` directives (and the above
function definition). The name of the function can be customized by providing an
argument to `--source`, but this will require modification of the shell completion provided.

Installation of shell completion can be achieved with:

```
sudo install -m644 contrib/completion.zsh /usr/share/zsh/site-functions/_jp
```

## Usage

Add a jump point (bookmark) with `jp -c <profile> <path>`. The path can be
absolute or relative, it will automatically be made absolute:

```sh
jump --create this-place .
jump -c jump ~/devel/go/src/github.com/cassava/jump
```

List jump points by simply running the command with no options or arguments.
