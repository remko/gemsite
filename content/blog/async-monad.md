---
title: "Flattening Callback Chains with Monad Do-Notation"
date: 2015-07-02
---

[A](http://journal.stuffwithstuff.com/2015/02/01/what-color-is-your-function/ "¹")
[few](https://code.facebook.com/posts/1661982097368498 "²")
[unrelated](https://groups.google.com/forum/#!msg/racket-users/RFlh0o6l3Ls/8InN7uz-Mv4J "³")
[topics](http://www.sunsetlakesoftware.com/2015/06/12/swift-2-error-handling-practice "⁴")
on my reading list made me want to randomly experiment with a few things.  I
wanted to understand monads a bit better, see how they applied to
callback-based asynchronous programming, and play around with macro
programming in a Lisp dialect. This is a partial log of the
theoretical-and-probably-not-directly-applicable-but-nevertheless-fun rabbit
hole I dived into.


## The Callback Pyramid Problem

JavaScript (and especially Node.js) uses lots of
asynchronous APIs. These are implemented by adding an extra callback parameter
to asynchronous methods, where the callback parameter is a function that
receives the result of the call when it's finished. 

```javascript
fetch("foo", (err, foo) => {
  // We have a 'foo' value. Do something with it.
  doSomethingWith(foo);
});
```

Although this is a very
simple and easy to understand way of implementing asynchronicity, the drawback is that it
becomes a bit ugly when chaining many of these calls:

```javascript
fetch("foo", (err, foo) => {
  doSomethingWith(foo);
  fetch("bar", (err, bar) => {
    doSomethingElse();
    fetch("baz", (err, baz) => {
      doSomethingWith(foo, baz);
    });
  });
});
```

Each asynchronous call needs to be nested inside the other, resulting in an 
ever-growing *callback pyramid of doom* or *callback hell* (and this trivial 
example doesn't even do error handling yet).

JavaScript offers libraries to help manage callbacks and flatten them out a bit
more, such as [async](https://github.com/caolan/async) or
[Promises](http://www.html5rocks.com/en/tutorials/es6/promises/).
Unfortunately, these libraries only tend to work well when the callback chains
are waterfalls, where each callback's result is only needed in the next
callback. When you need to combine values, you either get a pyramid again, or
some ugly code like this:

```javascript
let foo;
let baz;
fetch("foo")
  .then(result => { 
    foo = result; 
    doSomethingWith(foo);
    return fetch("bar");
  })
  .then(() => { 
    doSomethingElse();
    return fetch("baz");
  });
  .then(result => { 
    baz = result; 
    doSomethingWith(foo, baz);
  })
```

Yes, it's (relatively) flat, but the mutable variables at the top aren't very nice.

Ideally, you would be able to write asynchronous calls in a flat way, like you would
with synchronous code. For example, with some fictional syntax, you would be able to
write:

```javascript
async-do {
  let foo <- fetch("foo");
  doSomethingWith(foo);
  let bar <- fetch("bar");
  doSomethingElse();
  let baz <- fetch("baz");
  doSomethingWith(foo, baz);
}
```

Languages that have cooperative multithreading support (such as coroutines)
tend to use this to mimic synchronous code structure for asynchronous code, optionally 
by adding some syntactic sugar around it to make it even more convenient.
Examples include [C#'s
async/await](https://msdn.microsoft.com/en-us/library/hh191443.aspx "C# async/await"),
[Clojure's `core.async`](https://clojure.github.io/core.async/ "Clojure `core.async`"), and [Go's goroutines](https://tour.golang.org/concurrency/1 "Goroutines"). However, this introduces
an entirely different model of computation, usually not as easy to understand
as a simple callback mechanism.

So, I wanted to play with syntactic sugar to manage callbacks. Experimenting with syntax
is something that sounded like a nice job for a Lisp, so I used this as an excuse
to pick up [Scheme](http://schemers.org) and play around with this (well, 
[Racket](http://racket-lang.org) to be honest, but close enough).
The actual code used in this post can be found in [`async.rkt`](https://github.com/remko/toys/tree/scheme/async.rkt).


## Asyncs

Although we could use the same mechanism for callbacks as in JavaScript (where the 
callback is passed as the final parameter of a function), we use a slightly different
system: asynchronous functions return an *async*, a function that
takes a callback as parameter, executes the asynchronous logic, and then
feeds the result value (or error) to its callback parameter.

For example, suppose that we have a `fetch` procedure that does an asynchronous 
request for data. We let `fetch` return an async, such that we can call it with
a callback to print out the value:

```racket
; Create an async (i.e. a function) that will deliver the value of 
; "foo" to its callback argument
(define fetch-foo (fetch "foo"))

; Execute the request by passing a callback to the returned async; 
; the callback just prints out the result.
; We're ignoring errors for now.
(fetch-foo (λ (error value) 
              (display value) (newline)))
```

By using a `make-async` procedure that, similarly to JavaScript, returns an async
for a piece of code, we can create a dummy definition of a `fetch` function:

```racket
; Dummy version of a `fetch` procedure that returns an async for
; the value "<request>-value" (available after a few seconds)
(define (fetch key)
  ; The `make-async` argument is a piece of code that is executed, 
  ; and calls `resolve` or `reject` when the answer is ready or 
  ; has failed respectively.
  (make-async (λ (resolve reject)
                  (thread 
                    (λ ()
                      (sleep 2)
                      (resolve (string-append key "-value")))))))
```


Like with JavaScript, we can chain `fetch` calls into a lovely callback pyramid:

```racket
((fetch "foo") (λ (_ foo)
                  (do-something-with foo)
                  ((fetch "bar") (λ (_ bar)
                                  ; Don't need the value of bar
                                  (do-something-else)
                                  ((fetch "baz") (λ (_ baz)
                                                    (do-something-with foo baz)))))))
```


The pyramid looks even worse in Scheme than it does in JavaScript, so let's try to
get rid of it.


## The Async Monad

It turns out our async is actually a *monad*, a composable structure of
computations.  If you're unfamiliar with monads, you could have a look at [the WikiPedia page](https://en.wikipedia.org/wiki/Monad_%28functional_programming%29 "Monad · Wikipedia")
or [the Haskell page](https://wiki.haskell.org/Monad "Monad · Haskell"), but that probably won't
help you. Getting your head around monads is a challenge, so I won't attempt to
explain them here (I probably couldn't if I wanted to anyway). Instead, if you
have time, you should read the excellent [Learn You a Haskell for Great Good!](http://learnyouahaskell.com/ "Functors, Applicative Functors and Monoids · Learn You a Haskell for Great Good!"), which does a great job at explaining them step by step.
If you don't have time, that's fine too, you should be able to read on without
understanding anything about them.

Monads need to have 2 operations defined on them, `return` and `bind`, which we'll define for our asyncs.

`return` creates an async out of an ordinary value. Defining this one is easy: it just
creates an async that immediately resolves to the given value:

```racket
(define (async-return value)
  (make-async 
    (λ (resolve reject) (resolve value))))
```

`bind` creates a new async by feeding the succesful result 
of one async into a function returning another async. It may help to look at the
type definition of this operation in Haskell (where `bind` is called `>>=`):

```haskell
(>>=) :: Monad m => m a -> (a -> m b) -> m b
```

If we substitute the monad type `m` for a (fictional) `Async` type, this becomes:

```haskell
(>>=) :: Async a -> (a -> Async b) -> Async b
```

So, `bind` for asyncs is a function that takes an async of one type and a function 
that maps a value of that type onto another async, and returns a new combined async. 
Here is the definition of `async-bind`:

```racket
(define (async-bind async f)
  (make-async 
    (λ (resolve reject)
      (async (λ (async-error async-value)
                (if async-error
                    (reject async-error)
                    ((f async-value) (λ (f-error f-value) 
                                      (if f-error 
                                          (reject f-error)
                                          (resolve f-error))))))))))
```

If we use the `async-return` and `async-bind` operations, we can rewrite the example 
above like this:

```racket
(define fetch-foo-bar
  (async-bind (fetch "foo")
              (λ (foo)
                (async-bind (fetch "bar")
                            (λ (bar)
                              (async-return (do-something-with foo bar)))))))
```
  

Notice that:

- The callback doesn't have an `error` parameter anymore. The `async-bind` operation
  shortcuts the other callbacks when a failure occurs, and propagates the error
  to whoever uses the resulting async.
- We need to end the chain with a `async-return` call, because 
  `do-something-with` returns a regular value, whereas the second parameter of
  `async-bind` needs to be a function that returns a new async.

Using these operations doesn't look very practical in itself, 
it just creates even more pyramids. However, by copying some syntactic monad combining
sugar from Haskell, we can get a much nicer result.


## Do-notation

Because Haskell relies heavily on monads (mostly because you need it to do I/O
or model other side-effects), it has a special syntactic sugar to work with
monads, called the [*do-notation*](http://learnyouahaskell.com/a-fistful-of-monads#do-notation "do notation · Learn You a Haskell for Great Good!"). In Haskell, do-notation allows you to rewrite
    
```haskell
action1 >>= \x1 -> 
  action2 >>= \x2 ->
    action3 x1 x2
```

as

```haskell
do
  x1 <- action1
  x2 <- action2
  action3 x1 x2
```

We can introduce the same notation for our asyncs, flattening our `fetch-foo-bar` 
pyramid of `async-bind` calls into the following:

```racket
(async-do
  (<- foo (fetch "foo"))
  (<- bar (fetch "bar"))
  (async-return (do-something-with foo bar)))
```

Each argument of `async-do` is a call that creates an async, optionally embedded in
an `<-` arrow operator call to receive the value from the async and put it in a variable
(for use at a later statement in the `async-do`). Because this is just syntactic sugar
for combining asyncs, each line is only executed after the previous async has
been resolved.

Thanks to the power of Lisp, introducing the `async-do` syntactic sugar in Scheme can
be done with a simple macro:

```racket
(define-syntax async-do
  (syntax-rules (<-)
    ; The final statement is left as is
    [(_ e) e]

    ; A `var <- e` statement is rewritten into 
    ; `(async-bind e (λ (var) ...))`.
    [(_ (<- var e1) e2 ...)
      (async-bind e1 (λ (var) (async-do e2 ...)))]

    ; A regular statement `e` is rewritten into
    ; `(async-bind e (λ (_) ...))` (i.e. the value of the 
    ; async is ignored).
    [(_ e1 e2 ...) 
      (async-bind e1 (λ (_) (async-do e2 ...)))]))
```

This macro describes exactly how the do-notation is defined in Haskell.

Flattening the big pyramid from the beginning of this section with `async-do` becomes:

```racket
(define fetch-foo-bar-baz
  (async-do
    (<- foo (fetch "foo"))
    (async-return (do-something-with foo))
    (<- bar (fetch "bar"))
    (async-return (do-something-else))
    (<- baz (fetch "baz"))
    (async-return (do-something-with foo baz))))
```

The `async-return` calls, which ensure that every line in the `async-do` block is a
async, are still a bit annoying. Since we're in a dynamically typed language,
we can actually get rid of those by slightly changing the definition of the
`async-do` macro to automatically convert values into asyncs if necessary. 

```racket
(define-syntax async-do
  (syntax-rules (<-)
    [(_ e)
      (->async e)]
    [(_ (<- var e1) e2 ...)
      (async-bind (->async e1) (λ (var) (async-do e2 ...)))]
    [(_ e1 e2 ...)
      (async-bind (->async e1) (λ (_) (async-do e2 ...)))]))

; Convert a value to an async if it isn't one
(define-syntax ->async
  (syntax-rules ()
    [(_ e) 
      (let ([e-result e])
        (if (async? e-result)
          e-result
          (async-return e-result)))]))
```


The result of the pyramid above then finally becomes the clean, flat chain of callbacks
we set out to create:

```racket
(define fetch-foo-bar-baz
  (async-do
    (<- foo (fetch "foo"))
    (do-something-with foo)
    (<- bar (fetch "bar"))
    (do-something-else)
    (<- baz (fetch "baz"))
    (do-something-with foo baz)))
```


## Back to JavaScript?

So, we have a pretty clean way of doing asynchronous calls in Scheme, all based
on the well-understood foundations of monads. It would 
be useful if we would be able to bring this do-syntax into JavaScript too,
for example using the [Sweet.js](http://sweetjs.org) macro language. I haven't
had the courage to do this myself yet, because creating macros for JavaScript
is obviously more work due to the more complex syntactic structure of
JavaScript. Maybe something for a follow-up post?

In [the next post](/blog/scheme-monads "Next: 'More Fun with Monad Do-notation in Scheme'"), we'll go beyond the async monad with
the do-notation.
