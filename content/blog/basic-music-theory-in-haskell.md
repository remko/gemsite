---
title: "Basic Music Theory in Haskell"
date: 2008-06-19
---
While doing some spring cleaning around my hard disk, I found [a little Haskell program](https://github.com/remko/toys/blob/master/haskell/MusicTheoryBasics.hs "`MusicTheoryBasics.hs`") I wrote several years ago in an attempt to learn the basics of music theory. Now, I'm not a pro at writing Haskell, and I know even less about music theory, but I'm hoping that what I wrote down back then is a bit accurate. The program seems to summarize the basics quite consisely: by just having a glance at the program, I'm rediscovering some things I totally forgot about scales and chords.

For example, this is what it says about the sus4 chord:

```haskell
chordNotes Five = [(ScaleNote Major 1), (ScaleNote Major 5)]
chordNotes Sus4 = (chordNotes Five) ++ [(ScaleNote Major 4)]
```

So, sus4 is a power (5) chord (consisting of the first and the fifth of the major scale), added with the 4th note of the major scale. So, for Esus4, the program tells me:

```haskell
Main> scale2notes $ Scale (read "E") Major
[E,F#,G#,A,B,C#,D#]

Main> chord2notes $ Chord (read "E") Sus4
[E,A,B]
```

Something else I forgot is:
```haskell
intervals Ionian = [2,2,1,2,2,2,1]
intervals Major = Ionian
intervals scale = shift (intervals Ionian) (rank scale)
  where rank ...
```

So, every scale is really a shift of the major (well, any) scale, which is
actually called the Ionian scale.

This program might come in handy as a summary of music theory in case I forget
these things again 🙂.
