---
title: '"Beautiful Testing" XMPP Chapter'
date: 2009-05-03
---
Adam Goucher and Tim Riley (Director of QA at Mozilla) [announced](http://adam.goucher.ca/?p=684 "'Beautiful Testing is a Go!' · Adam Goucher") a few months ago that they are putting together a [Beautiful Testing book](http://oreilly.com/catalog/9780596159818 "Beautiful Testing · O'Reilly") for O’Reilly. I took the opportunity to write a chapter about testing in the context of XMPP (more specifically, about testing protocol implementations in [Swift](https://swift.im)), and just submitted the final draft for technical review. The book is expected to be released this August.

Although there are many types of testing being done in the XMPP world, the
chapter focuses on the beauty of testing the functionality of XMPP protocol
implementations. After a brief introduction on XMPP, it starts out with a
description of unit testing simple IQ request/response protocols, and  then
gradually moves on to higher-level testing of more complex, multi-stage
protocols such as session initialization. As you might expect from a developer
like me, the chapter is quite heavy on the (C++) code, but I’m told it
compensates for the rest of the book 😉

As with all other books in the O’Reilly “Beautiful” series (which started with
Beautiful Code, but has since been followed up by Beautiful Architecture,
Beautiful Teams, Beautiful Security, and Beautiful Data), all proceeds of the
book go to charity, in this case [“Nothing But Nets”](http://www.nothingbutnets.net/ "Nothing But Nets") (which provides mosquito
netting to malaria infested areas of Africa). This means that I can plug this
book as much as I want, and still have the feeling I’m actually doing a noble,
unselfish thing. (contrary to when I casually mention that you can buy our book
["XMPP: The Definitive Guide"](http://oreilly.com/catalog/9780596521264 "XMPP: The Definitive Guide · O'Reilly") at very sharp prices these days). Some time after
the book’s release this summer, I will even make a free version of the chapter
available here, so check back soon!

