syncthing
=========

[![Latest Build](http://img.shields.io/jenkins/s/http/build.syncthing.net/syncthing.svg?style=flat-square)](http://build.syncthing.net/job/syncthing/lastBuild/)
[![API Documentation](http://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](http://godoc.org/github.com/syncthing/syncthing)
[![MPLv2 License](http://img.shields.io/badge/license-MPLv2-blue.svg?style=flat-square)](https://www.mozilla.org/MPL/2.0/)

This is the `syncthing` project which pursues the following goals:

 1. Define a protocol for synchronization of a folder between a number of
    collaborating devices. This protocol should be well defined, unambiguous,
    easily understood, free to use, efficient, secure and language neutral.
    This is called the [Block Exchange
    Protocol](https://github.com/syncthing/specs/blob/master/BEPv1.md).

 2. Provide the reference implementation to demonstrate the usability of
    said protocol. This is the `syncthing` utility. We hope that
    alternative, compatible implementations of the protocol will arrise.

The two are evolving together; the protocol is not to be considered
stable until syncthing 1.0 is released, at which point it is locked down
for incompatible changes.

Getting Started
---------------

Take a look at the [getting started
guide](https://github.com/syncthing/syncthing/wiki/Getting-Started).

There are a few examples for keeping syncthing running in the background
on your system in [the etc directory](https://github.com/syncthing/syncthing/blob/master/etc).

There is an IRC channel, `#syncthing` on Freenode, for talking directly
to developers and users.

Building
--------

Building Syncthing from source is easy, and there's a
[guide](https://github.com/syncthing/syncthing/wiki/Building).
that describes it for both Unix and Windows systems.

Signed Releases
---------------

As of v0.10.15 and onwards, git tags and release binaries are GPG signed
with the key D26E6ED000654A3E (see https://syncthing.net/security.html).
For release binaries, MD5 and SHA1 checksums are calculated and signed,
available in the md5sum.txt.asc and sha1sum.txt.asc files.

Documentation
=============

The [syncthing
documentation](https://github.com/syncthing/syncthing/wiki/) is on the
Github wiki.

All code is licensed under the
[MPLv2 License](https://github.com/syncthing/syncthing/blob/master/LICENSE).
