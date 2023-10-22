---
title: "Ken Burns Effect Slideshows with FFMpeg"
date: 2017-08-06
featured: true
---
One of the first things that impressed me about Mac OS X when I first saw it
was its screensaver. Instead of just showing a simple slideshow of your
pictures, it actually used a [Ken Burns](https://en.wikipedia.org/wiki/Ken_Burns_effect "Ken Burns effect -- Wikipedia") panning and zooming
effect with a fancy fading transition to make the otherwise static pictures
really come to life. It always sounded like a fun project to create a standalone
tool to create slideshow movies that used this effect, with full control over
where and how much pictures should be zoomed. It turns out you don't really
need a new tool: you can get the same result using just
[FFMpeg](https://ffmpeg.org). In this post, I'll walk through the steps of
creating slideshows using the Ken Burns effect. 

Oh, and cats. There will be cats!


## Zooming and Panning

Let's start with the basics: applying a zoom effect to an image. FFMpeg
has a [`zoompan`](https://ffmpeg.org/ffmpeg-filters.html#zoompan "FFMpeg `zoompan` filter") 
filter that does exactly this. For example, to zoom in on an image:
```
ffmpeg -i in.jpg 
  -filter_complex 
    "zoompan=z='zoom+0.002':d=25*4:s=1280x800" 
  -pix_fmt yuv420p -c:v libx264 out.mp4
```

The `zoompan` filter takes a `z` expression that is evaluated for each frame to
know how far should be zoomed in. Here, we zoom in linearly 20% by adding
`0.002` to the previous zoom value on each frame. The other parameters we
specified are the duration of the effect (4 seconds at the default 25 frames
per second), and the output size for the resulting video. Finally, this is
output to an H.264 movie (using the standard YUV420p pixel format). The result
looks like this:

![Zoom](/blog/ken-burns-ffmpeg/zoom.gif)

If we only change the `z` parameter, the image will zoom in to the top left.
By also specifying the `x` and `y` parameters (which both default to 0), we
can zoom in to other parts of the picture:

- If we want to zoom to the *right* part of the picture, we specify `x` to
  be `iw-iw/zoom` (the input width, minus the width of the zoomed frame)
- If we want to zoom to the *bottom* part of the picture, we specify `y`
  to be `ih-ih/zoom`.

To zoom *out* instead of zooming in, we need to initialize the first frame to
be completely zoomed in, and decrease the zoom on every frame
(e.g. `z='if(eq(on,1),1.2,zoom-0.002)` for zooming out from 20%).

Putting it all together, zooming out from the bottom right would be done
with the following filter:

```
zoompan=
  x='iw-iw/zoom':
  y='ih-ih/zoom':
  z='if(eq(on,1),1.2,zoom-0.002)':
  d=25*4:s=1280x800
```


## Handling different aspect ratios

The `zoompan` filter scales the input to the specified output dimensions.
This means that, if the input picture has a different aspect ratio than the
output video (which will certainly be the case for portrait pictures), one of the 
dimensions will be stretched. 

### Cropping

A first solution to this problem is to zoom/pan to a video of the same aspect
ratio as the input picture, and then simply cut away the leftover parts using 
a [`crop`](https://ffmpeg.org/ffmpeg-filters.html#crop "FFMpeg `crop` filter") filter. For example,
for a portrait picture, the resulting filterchain would be:

```
zoompan=
  z='zoom+0.002':d=25*4:s=1280x2048,
crop=
  w=1280:h=800:x='(iw-ow)/2':y='(ih-oh)/2' 
```

![Crop](/blog/ken-burns-ffmpeg/crop.gif)

### Padding

Cropping works when the aspect ratio of the video doesn't differ much
from the input picture. For portrait pictures, however, this typically
cuts away too much, leaving little of the original picture. A second solution to the problem is to first add
padding to the picture to make it match the aspect ratio of the video. 
This can be done by putting a
[`pad`](https://ffmpeg.org/ffmpeg-filters.html#pad-1 "FFMpeg `pad` filter") filter before the `zoompan`
filter. For example, for a 3750×6000 picture, we first add padding to the sides
before passing it to the `zoompan` filter:

```
pad=
  w=9600:h=6000:x='(ow-iw)/2':y='(oh-ih)/2',
zoompan=
  z='zoom+0.002':d=25*4:s=1280x800
```

![Pad](/blog/ken-burns-ffmpeg/pad.gif)

### Panning

If you don't like the black bars around the image, but still want to show
most of your image, another alternative is to do extra panning while zooming.
For example, when you are zooming in on a portrait picture, pan from the bottom
to the image to the top of the image while applying the zoom effect.

This effect is a bit more complex to accomplish, and requires multiple
changes to the filterchain:

- First, the picture needs to be padded to match the output aspect ratio,
  exactly as in the previous section
- The initial zoom factor needs to be adjusted to take the extra padding
  into account. For example, zooming in 20% on the padded 3750×6000
  portrait above would need to start with an initial zoom factor of 2.56 
  ((1280/800)/(3750/6000)) instead of the default 1
- Depending on whether the padding is applied horizontally or vertically, 
  an extra offset needs to be added to the `x` or `y` component, while
  the other component needs to change dynamically to achieve the pan effect.
  For zooming in on the example portrait picture above, this would yield the offset
  `x='(iw-(3750/6000)*ih)/2'` and the 
  panning `y='(1-on/(25*4))*(ih-ih/zoom)'` (4 seconds at 25 fps).

All together, zooming in 20% on the top left of the portrait picture above while 
panning from bottom to top:

```
pad=
  w=9600:h=6000:
  x='(ow-iw)/2':y='(oh-ih)/2',
zoompan=
  x='(iw-0.625*ih)/2':
  y='(1-on/(25*4))*(ih-ih/zoom)':
  z='if(eq(on,1),2.56,zoom+0.002)':
  d=25*4:s=1280x800
```

![Pan](/blog/ken-burns-ffmpeg/pan.gif)

To zoom in to the bottom right, while panning from top to bottom, the extra `x`
expression we added in the first section needs to be scaled to take into
account the padding:

```
pad=
  w=9600:h=6000:
  x='(ow-iw)/2':y='(oh-ih)/2',
zoompan=
  x='(iw-0.625*ih)/2+ih*0.625-ih*1.6/zoom':
  y='(on/(25*4))*(ih-ih/zoom)':
  z='if(eq(on,1),2.56,zoom+0.002)':
  d=25*4:s=1280x800
```

Note that the more distance the pan needs to cover, the higher the frame rate
has to be to maintain a smooth effect.

## Adding transitions

A typical effect applied in slideshow movies is to fade between pictures
to get a smooth transition. FFMpeg doesn't have a built-in
crossfade filter for video, but there are different ways to get this
effect with built-in filters.

One way is to fade out the alpha channel at the end of one picture,
fade in the alpha channel at the beginning of the other, make the
beginning and end fade of both pictures overlap, and overlay the result
on a black video. For example, the entire FFMpeg command to have a 1
second fade between 2 pictures is:

```
ffmpeg -i in1.jpg -i in2.jpg 
  -filter_complex 
    "color=c=black:r=60:size=1280x800:d=7.0[black];
    [0:v]format=pix_fmts=yuva420p,zoompan=d=25*4:
    s=1280x800,fade=t=out:st=3.0:d=1.0:alpha=1,
    setpts=PTS-STARTPTS[v0];[1:v]format=
    pix_fmts=yuva420p,zoompan=d=25*4:s=1280x800,
    fade=t=in:st=0:d=1.0:alpha=1,setpts=PTS-
    STARTPTS+3.0/TB[v1];[black][v0]overlay[ov0];
    [ov0][v1]overlay=format=yuv420"
  -c:v libx264 out.mp4
```

It might be a bit easier to understand if we add some spacing to the
filter graph:

```
  color=
    c=black:
    size=1280x800:
    d=7.0
[black];

[0:v]
  format=pix_fmts=yuva420p,
  zoompan=d=25*4:s=1280x800,
  fade=
    t=out:
    st=3.0:
    d=1.0:
    alpha=1
[v0];

[1:v]
  format=pix_fmts=yuva420p,
  zoompan=d=25*4:s=1280x800,
  fade=
    t=in:
    st=0:
    d=1.0:
    alpha=1,
  setpts=PTS-STARTPTS+3.0/TB
[v1];

[black][v0]
  overlay
[ov0];

[ov0][v1]
  overlay=format=yuv420
```

Note that the parts between square brackets are the inputs and output names of
the filterchains. The first `color` filter is an input source of black video
for the duration of the entire video. Then follow the 2 filterchains of the
images we pass as input to FFMpeg. These take as input the video streams
passed as `-i` parameters, add an alpha channel using the `format` filter,
apply the `fade` to alpha filter (the fade out starts at 3 seconds, the fade
in at 0 seconds), and then the `setpts` filter is used to set the timestamp
of the second picture to start at 3 seconds (overlapping 1 second with the
first picture). Finally come a series of `overlay` filter applications,
overlaying the first picture on the black video, and the second picture on
top of that result.

You can optionally add an extra `fade` filter on the first and the last
picture to start and end your presentation with a fade effect.

The result for applying all fades on 2 pictures looks like this:

![Fade effects](/blog/ken-burns-ffmpeg/fade.gif)

## Looping

Instead of starting and ending your slideshow movie with a fade effect, you
can adapt it so it is infinitely loopable, transitioning smoothly from the 
last picture to the first picture.

To achieve this, all you need to do is to use a duplicate of the filterchain
of the first image as the last image, and trim your video to skip the duration
of the fade at the beginning, and stop after the fade in of the last picture.
For example, for 3 pictures with a 1 second fade and a total of 4 seconds per 
picture, this would be:

```
ffmpeg -i 1.jpg -i 2.jpg -i 3.jpg 
  -filter_complex 
    "[black]…;[0:v]…;[1:v]…;[2:v]…;[0:v]…"
  -ss 1 -t 9 
  -c:v libx264 out.mp4
```

![Loop](/blog/ken-burns-ffmpeg/loop.gif)

## All together now

Combining all the filters above quickly becomes very complex. For example,
for a presentation of just 3 pictures, the entire FFMpeg command would be:

```
ffmpeg -i 1.jpg -i 2.jpg -i 3.jpg 
  -filter_complex "color=c=black:r=60:size=1280x800:d=10[black];[0:v]format=pix_fmts=yuva420p,crop=w=2*floor(iw/2):h=2*floor(ih/2),zoompan=z='if(eq(on,1),1,zoom+0.000417)':x='0':y='ih-ih/zoom':fps=60:d=60*4:s=1280x800,crop=w=1280:h=800:x='(iw-ow)/2':y='(ih-oh)/2',fade=t=in:st=0:d=1:alpha=0,fade=t=out:st=3:d=1:alpha=1,setpts=PTS-STARTPTS[v0];[1:v]format=pix_fmts=yuva420p,crop=w=2*floor(iw/2):h=2*floor(ih/2),pad=w=9600:h=6000:x='(ow-iw)/2':y='(oh-ih)/2',zoompan=z='if(eq(on,1),1,zoom+0.000417)':x='0':y='0':fps=60:d=60*4:s=1280x800,fade=t=in:st=0:d=1:alpha=1,fade=t=out:st=3:d=1:alpha=1,setpts=PTS-STARTPTS+1*3/TB[v1];[2:v]format=pix_fmts=yuva420p,crop=w=2*floor(iw/2):h=2*floor(ih/2),zoompan=z='if(eq(on,1),1,zoom+0.000417)':x='0':y='0':fps=60:d=60*4:s=1600x800,crop=w=1280:h=800:x='(iw-ow)/2':y='(ih-oh)/2',fade=t=in:st=0:d=1:alpha=1,fade=t=out:st=3:d=1:alpha=0,setpts=PTS-STARTPTS+2*3/TB[v2];[black][v0]overlay[ov0];[ov0][v1]overlay[ov1];[ov1][v2]overlay=format=yuv420"
  -c:v libx264 out.mp4
```

So, you'll probably want to script all this. I created [a `kburns.rb` script](https://github.com/remko/kburns "`kburns.rb`") that does exactly this. You're welcome
to try it out, and feel free to modify it as you want.

