# deck - hands-off package manager

There is no shortage of traditional package management tools for linux. Every distro has 
a different one - rpm, apt, dpkg, pacman, yum, emerge etc. They all take a pre-built 
package archive, expand it on top of your root fs and record the files they just added 
to a database. When you remove a package, the package manager looks up the filenames in 
the database and removes them. The creation of the initial package archive is hidden from
the enduser and handled by the distros package maintainers or build scripts in the ports
system.

But what if all you've got is a source tarrball and a makefile, like when you build
Linux From Scratch or work on a custom distribution? You can build and install it 
with `make install`, but usually there is no `make uninstall`. The LFS project 
privides some hints on how to [add package management to an LFS system](http://www.linuxfromscratch.org/lfs/view/development/chapter06/pkgmgt.html),
but all of the methods require a lot of extra work and are not 100% proof.

This small project tries to address this problem:

* install a package with `./configure --prefix=/usr && make && make install` 
* run `deck scan` to see what files have been installed, modified or deleted
* run `deck commit` to remember the changes or `deck reset` to discard
* run `deck uninstall` to remove a previously installed package

## how it works

`deck` was built with two assumptions:

1. modern hardware is fast: you can walk the filesystem and calculate a hash for 
   every file in reasonable time. On my old _Sandy Bridge Core i5_ laptop with an
   oldish _SSD_ drive hashing the full system (~4gb) takes about a minute.
2. storage is cheap. you can keep a backup copy of every file and restore it if
   it was modified

`deck` keeps a record of the file's metadata and a copy of the contents for every file
you add to the database. Then when you scan the filesystem for changes it compares it
with the previously recorded state and reports any files that have been added, modified
or deleted.

`deck` is a written in *Go* and *statically linked*. There are no external dependencies, 
not even libc. You can drop it anywhere in your *$PATH* and start using right away.

## example session

	root@warthog ~ # tar xf src/rfkill-0.5.tar.xz
	root@warthog ~/rfkill-0.5 # make
	 CC   rfkill.o
	 GEN  version.c
	 CC   version.o
	 CC   rfkill
	root@warthog ~/rfkill-0.5 # make PREFIX=/usr install
	 GZIP rfkill.8
	 INST rfkill
	 INST rfkill.8
	root@warthog ~/rfkill-0.5 # cd .. 
	root@warthog ~ # deck scan
	New files :

		 /usr/bin/rfkill
		 /usr/share/man/man8/rfkill.8.gz

	root@warthog ~ # deck pick /usr/bin/rfkill /usr/share/man/man8/rfkill.8.gz
	root@warthog ~ # deck commit -p rfkill -v 0.5
	root@warthog ~ # deck show rfkill
	/usr/bin/rfkill
	/usr/share/man/man8/rfkill.8.gz
	root@warthog ~/rfkill-0.5 # deck uninstall rfkill
	rm /usr/bin/rfkill
	rm /usr/share/man/man8/rfkill.8.gz
	root@warthog ~ #

# config file

deck looks for the config file in `/etc/deckrc` or `${HOME}/.deckrc`. Config file location can be 
overriden with the global `--config` option

* `root`   is the top directory for filesystem scans
* `data`   is the directory where it keeps its database and a copy of the tracked files
* `prune`  is a list of regular expressions. If a directory name matches one of the 
   expressions, deck will skip it
* `ignore` - same as prune but for files
* `git` -  if there is a git repsitory at the specified `root` and `git` is set to `true`,
   deck will ignore all files tracked by git

there is an example config to start with in `deckrc.example`

# installation

A prebuilt static binary can be obtained in the [releases section](https://github.com/pampa/deck/releases) on github.
Download the most recent release, unpack it with xz -d, set the executable bit and put it in your *$PATH*

## build from source

To build from source you will need a working copy of the [Go compiler](https://golang.org/doc/install)
The following instructions will build a statically linked binary:

	~ $ git clone https://github.com/pampa/deck.git
	~ $ cd deck
	~/deck $ go get -v -d
	~/deck $ make deck

You can also use the `go get` method, but the resulting binary will be dynamically linked against glibc

	go get -v github.com/pampa/deck

## cross compile for a different platform

Deck can be easily cross compiled for a different platform and architecture
	
	~ $ git clone https://github.com/pampa/deck.git
	~ $ cd deck
	~/deck $ go get -v -d
	~/deck $ GOOS=linux GOARCH=arm go build -v

This will build a statically linked binary for the target platform. For a list of supported target platforms see https://golang.org/doc/install/source#environment

# usage

	deck [global options] command [command options] [arguments...]
   
## commands

	scan, s              scan the filesystem for changes
	pick, p              pick file for further processing
	unpick, u            unpick file
	commit               commit picked files to index, adding package and version tags
	list, l              list all packages in index
	show, o              show package contents
	remove, rm           remove file from index
	reset                reset file to its previous state
	uninstall            uninstall package	
	which, w, who, what  show which package a file belongs to
	doctor, doc, d       run database sanity checks
	help, h              Shows a list of commands or help for one command

## global options

	--config value, -c value  config file to use instead of /etc/deckrc or $HOME/.deckrc
	--debug, -d               print debug info to stderr
	--help, -h                show help
	--version, -v             print the version
   
## deck scan - scan the filesystem for changes
	
	deck scan [command options] [arguments...]

   	--hash, -s  use sha1 to compare files
   	--pick, -p  pick new files
   
## deck pick - pick file for further processing
	
	deck pick [arguments...]

## deck commit - commit picked files to index, adding package and version tags

	deck commit [command options] [arguments...]

	--package value, -p value  package name
	--version value, -v value  package version
   
## deck list - list all packages in index

	deck list [command options] [arguments...]
	
	--version, -v  do not print version number
   
## deck show - show package contents

	deck show [arguments...]
   
## deck remove - remove file from index

	deck remove [arguments...]

## deck reset - reset file to its previous state

	deck reset [arguments...]

## deck uninstall - uninstall package

	deck uninstall [arguments...]

## deck which - show which package a file belongs to
	
	deck which [arguments...]

## deck doctor - run database sanity checks
	
	deck doctor [arguments...]
