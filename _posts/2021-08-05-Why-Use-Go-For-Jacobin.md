layout: post
title: "Why use Go for Jacobin?"
date: 2021-08-05 18:00:00 -0800
categories: Go 

Choosing a Language
I have spent the last eight months researching the JVM--reading the docs and articles and doing exploratory coding in various languages with which to write the Jacobin JVM. My requirements for the implementation language are simple enough: it must have decent tools and a viable ecosystem, it must compile to native code on the three major platforms (Windows, Mac, and Linux), and it must have built-in garbage collection (GC). The latter requirement is important. The JVM performs garbage collection, but I don't want to write a garbage collector. They are exceedingly difficult tools to write and, especially, to debug. By using a language that does its own GC, a huge amount of work has been removed from the project.

Three languages meet my requirements: Dart, Swift, and Go. I've written several thousand lines of code in the first two and have eliminated them from consideration. Here is why. Dart is a lovely language, but it's slow (even when compiled to binaries), its ecosystem is wanting, and the kind of threading it does is a poor match to the JVM. The problem with the ecosystem is exemplified by the nearly complete absence books on the language since Dart 2.0 came out a few years ago. Almost all written tutorials are way out of date. Those that are current focus, without exception, on Flutter--the UI toolkit that dominates the use cases for Dart. As a result, it's not easy to learn Dart in depth unless you want to focus primarily on Flutter. The Dart team should really address this. As to the threading model, it is based entirely on single-channel message passing: there is no shared memory. The JVM must perforce share memory between threads and so even if Dart were faster and the docs were up-to-date, it would not meet my needs.

Swift is a truly beautiful language. It's rich in features and has a lot of the type-checking and code safety rules of Rust, but without the endless head-banging that Rust entails. I would have loved to write the JVM in Swift, but it has several drawbacks: it doesn't run on Windows and its libraries are intimately tied to the Mac. Let me clarify. There is an official version of Swift for Windows, but it's maintained entirely by a single engineer at Google. There are effectively no docs for this version and the installation instructions don't work no matter how much tweaking and configuration I have done. The second problem is that while Swift is trying to become a language that works beyond just Apple platforms (for example, it runs fine on Linux), this worthy goal is far from especially when it comes to libraries. Consider that the equivalent of libzip (which is a core library in most languges--it is used to compress/decompress data using the zip format) is maintained by a third party on Github on a project that has at present 22 stars. The collections library has at most a handful of basic data structures, etc. Unless I want to write many of these libraries myself--which I have no desire to do--I am forced down the same road as Node developers: grabbing bits of functionality here and there from different contributors, many of which have unknown code quality. The alternative is to use Apple's Cocoa frameworks on the Mac, which would make my project Mac-only. In sum, until Swift grows its non-Mac ecosystem, it's not a viable option for this project--much to my chagrin.

This leaves Go, which is an easy-to-learn language that runs well on the major platforms and has a flourishing set of libraries, many of which are maintained by core Go developers. While it checks all the boxes, it presents its own challenges. For example, it's the only one of the languages that is not object-oriented and the transition from thinking in objects (after all, Java is my home language, so to speak) to using an imperative style of coding requires some rewiring of how I approach problems. In addition, the standard Go tools have weaknesses. For example, the testing framework is minimal--there is nothing like JUnit in terms of range of features. In the language itself, return values for errors and the lack of generics both feel a little crude, especially to someone coming from Java. Nonetheless, it looks like the best option for my project.

There was one other language candidate: Java. That is, write a JVM that runs on the JVM. I don't find this interesting at all. The code for the JVM is currently mostly written in Java and I'll be consulting it frequently--so what would I do then? Cut and paste? Rewrite the code in my preferred style? It's hard to see how that's an advantage.