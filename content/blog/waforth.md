---
title: "A Dynamic Forth Compiler for WebAssembly"
date: 2018-05-24
featured: true
---
In yet another 'probably-useless-but-interesting' hobby project, I wrote a Forth compiler and interpreter targeting WebAssembly.
It's written entirely in [WebAssembly](https://webassembly.org), and comes with a compiler
that dynamically emits WebAssembly code on the fly. The entire system (including 80% of all
core words) fits into a 10k (5k gzipped) WebAssembly module. You can try out [the WAForth interactive console](https://mko.re/waforth "WAForth Interactive Console"), or
grab the code from [GitHub](https://github.com/remko/waforth "WAForth GitHub page").

What follows are some notes on the design, and some initial crude speed benchmarks.

![WAForth Interactive Console](/blog/waforth/console.gif "WAForth Interactive Console")


> ℹ️ Note (2022-08-19): This post is relatively old. Since the time of writing, a lot has been added
> to WAForth, including design changes, the implementation of all the ANS Core words and most 
> ANS Core Extension words, the addition of a JavaScript interface, a standalone version, ... 
>
> For a more up-to-date view of the project, check out [the WAForth GitHub page](https://github.com/remko/waforth "WAForth GitHub page").

> ℹ️ Note (2023-02-25): Don't like reading? Have a look at [my FOSDEM'23 talk on WAForth](https://www.youtube.com/watch?v=QqW39jElFhA "FOSDEM'23 WAForth Talk -- YouTube").

## Forth

[Forth](https://en.wikipedia.org/wiki/Forth_%28programming_language%29) is a
low-level, minimalistic stack-based programming language. 

Forth typically comes in the
form of an interactive interpreter, where you can type in your commands. For example,
taking the sum of 2 numbers and printing the result:

```
2 4 + .             6 ok
```

Forth environments also have a compiler built-in, allowing you to define new 'words' 
by typing their definition straight from within the interpretter:

```
: QUADRUPLE  4 * ;
```

which you can then immediately invoke

```
2 QUADRUPLE .       8 ok
```

Not unlike Lisps, you can customize Forth's compiler, add new control flow
constructs, and even switch back and forth between the interpreter and the compiler
while compiling.

Because of its minimalism, Forth environments can be easily ported to new
instruction sets, making them popular in embedded systems. To learn a bit more
about this language (and about WebAssembly), I wanted to try creating an
implementation for WebAssembly -- not exactly an embedded instruction set, but
an instruction set nonetheless.

## Design

WAForth is (almost) entirely written in WebAssembly. The only parts for which
it relies on external (JavaScript) code is the dynamic loader (which isn't
available
([yet?](https://webassembly.org/docs/future-features/#platform-independent-just-in-time-jit-compilation "Platform-independent JIT compilation -- WebAssembly Future Features"))
in WebAssembly), and the I/O primitives to read and write a character.

I got a lot of inspiration from [jonesforth](http://git.annexia.org/?p=jonesforth.git;a=tree), a
minimal x86 assembly Forth system, written in the form of a tutorial.

### The Macro Assembler

> Update (11/2019): WAForth no longer uses a custom macro assembler; the core is now written
> entirely in raw WebAssembly.

The WAForth core is written as [a single module](https://github.com/remko/waforth/blob/master/src/waforth.wat "WAForth Core WebAssembly module") in WebAssembly's [text format](https://webassembly.github.io/spec/core/text/index.html "WebAssembly Text Format"). The 
text format isn't really meant for writing code in, so it has no facilities like a real assembler
(e.g. constant definitions, macro expansion, ...) However, since the text format uses S-expressions,
you can do some small tweaks to make it loadable in a Lisp-style system, and use it to extend
it with macros.

So, I added some Scheme ([Racket](https://racket-lang.org)) macros to the module definition, 
and implemented a mini assembler to print out the resulting s-expressions in a compliant WebAssembly format.

The result is something that is almost exactly like a standard WebAssembly
text format module, but sprinkled with some macros for convenience.

### The Interpreter

The interpreter runs a loop that processes commands, and switches to and from
compiler mode. 

Contrary to some other Forth systems, WAForth doesn't use direct threading 
for executing code, where generated code is interleaved with data, and the
program jumps between these pieces of code. WebAssembly doesn't allow
unstructured jumps, let alone dynamic jumps. Instead, WAForth uses
subroutine threading, where each word is
implemented as a single WebAssembly function, and the system uses calls and
indirect calls (see below) to execute words.


### The Compiler

While in compile mode for a word, the compiler generates WebAssembly instructions in
binary format (as there is no assembler infrastructure in the browser). Because WebAssembly
[doesn't support JIT compilation yet](https://webassembly.org/docs/future-features/#platform-independent-just-in-time-jit-compilation "Platform-independent JIT compilation -- WebAssembly Future Features"), a finished word is bundled into a separate binary WebAssembly module, and
sent to the loader, which dynamically loads it and registers it in a shared 
[function table](https://webassembly.github.io/spec/core/valid/modules.html#tables "WebAssembly Tables") at the
next offset, which in turn is recorded in the word dictionary. 

Because words reside in different modules, all calls to and from the words need to happen as
indirect `call_indirect` calls through the shared function table. This of course introduces
some overhead.

As WebAssembly doesn't support unstructured jumps, control flow words (`IF/ELSE/THEN`, 
`LOOP`, `REPEAT`, ...) can't be implemented in terms of more basic words, unlike in jonesforth.
However, since Forth only requires structured jumps, the compiler can easily be implemented 
using the loop and branch instructions available in WebAssembly.

Finally, the compiler adds minimal debug information about the compiled word in
the [name section](https://github.com/WebAssembly/design/blob/master/BinaryEncoding.md#name-section "WebAssembly Name Section"), making it easier for doing some debugging in the browser.

![Debugger view of a compiled word](/blog/waforth/debugger.png "Debugger view of a compiled word")


### The Loader

The loader is a small bit of JavaScript that uses the [WebAssembly JavaScript API](https://webassembly.github.io/spec/js-api/index.html) to dynamically load a compiled word (in the form of a WebAssembly module), and ensuring that the shared function table is large enough for the module to
register itself.

### The Shell

There's a small shell around the WebAssembly core to interface it with JavaScript.
The shell is [a simple class](https://github.com/remko/waforth/blob/master/src/shell/WAForth.js "WAForth JavaScript wrapper") 
that loads the WebAssembly code in the browser, 
provides the loader and the I/O primitives to the WebAssembly module to read and write characters to a terminal. On the other end, it provides a `run()` function to execute a fragment of Forth code.

To tie everything together into an interactive system, there's a small
console-based interface around this shell to type Forth code, which you can see
in action [here](https://mko.re/waforth "WAForth Interactive Console").

![WAForth Console](/blog/waforth/console.gif "WAForth Console")

## Benchmarks

> ⚠️ Note (2022-08-19): I wouldn't trust the results below any more:
>
> - A lot has changed in browser WebAssembly implementation, both good (speedups) and 
>   bad (Spectre mitigations). I have no idea in which direction this takes the recorded times.
> - A lot has changed in WAForth itself in terms of design (including optimizations that yield 40%
>   speedups on this toy benchmark, but also changes that slowed down)
> - For GForth times, I did not look in details at any flags. For all I know, Gforth could 
>   generate even faster code by tweaking some flags.
> - The benchmark is a very toy benchmark. I have no idea how representative it is.


Although I didn't focus on performance while writing WAForth, I still wanted to have an 
estimate of how fast it went. To get a crude idea, I ran an implementation of the
[Sieve of Eratosthenes](https://en.wikipedia.org/wiki/Sieve_of_Eratosthenes). I let the algorithm compute
all the primes up to 90 million on my MacBook Pro 2017 (3.5Ghz Intel Core i7) running Firefox 60.0.1, and timed different systems:
- **WAForth**: The sieve algorithm, [written in Forth](https://rosettacode.org/wiki/Sieve_of_Eratosthenes#Forth "Sieve of Eratosthenes in Forth"),  running in WAForth. The words are compiled as separate WebAssembly modules, and all calls to and from the words are indirect (as described above).
- **WAForth Direct Calls**: The sieve algorithm, as compiled by WAForth, but inserted directly in the WAForth core, substituting all indirect call instructions from the previous version for direct calls. This measurement gives an indication of the overhead of indirect jumps.
- **Gforth**: The sieve algorithm running in [Gforth Fast 0.7.3](https://www.gnu.org/software/gforth/ "GForth"), a native, high-performance Forth. 
- **JS-Forth**: The sieve algorithm running in [JS-Forth 0.5200804171342](http://www.forthfreak.net/jsforth.html "JS-Forth"), a JavaScript implementation of Forth.
- **Vanilla WebAssembly**: A [straight WebAssembly implementation of the algorithm](https://github.com/remko/waforth/blob/master/tests/benchmarks/sieve-vanilla/sieve-vanilla.wat "Sieve of Eratosthenes in hand-written WebAssembly").   This serves as an upper bound of how fast the algorithm can run on WebAssembly.

![Benchmarks](/blog/waforth/benchmarks.png "Sieve Benchmarks")

Some observations:

- Not surprisingly, JS-Forth comes out the slowest. It also runs into memory problems when trying to increase the 90 million limit. The other implementations have no problem (requiring only 1 byte per candidate prime).
- The indirect calls in WAForth cause ±50% overhead. Some of this overhead can be reduced when
  WebAssembly starts supporting [mutable globals](https://github.com/WebAssembly/mutable-global "WebAssembly mutable globals"), as
  this will require less indirect calls for operations like loops and jumps.
- WAForth is 2× slower than the high-performance native Gforth
- The Vanilla WebAssembly version of the Sieve algorithm is much faster than the rest. Contrary to the WAForth version, this version 
  doesn't need to go to memory for every piece of control flow, causing massive speed gains. This
  is especially noticeable when increasing the number of primes: for all primes less than 500 million,
  the vanilla WebAssembly version is up to 8 times faster than the WAForth one.

WAForth is still experimental, and [lacks support for a few of the ANS core words](https://github.com/remko/waforth/issues/4 "'Implement all ANS Core Words' -- WAForth GitHub Issues") (although adding support for these shouldn't be too much work). I also didn't spend any time profiling performance to see if there are any
low-hanging fruit optimizations to be done (although I think most optimizations would complicate
the compiler too much at this point).

