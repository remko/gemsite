---
title: "HAProxy Alerts with WebHooks"
date: 2016-08-16
---
I want to be notified immediately when one of the backend servers behind my
[HAProxy](http://www.haproxy.org) instance goes down. HAProxy offers alerting
functionality, but only via SMTP: when a backend goes down it sends an email to
a list of recipients via an SMTP server you provide. Unfortunately, email isn't
an ideal mechanism for real-time notification, and I don't have an SMTP server
accessible from my HAProxy instance. 

In this post, I'll show my setup of HAProxy posting its alerts to a Slack channel and the Pushover push notification service, using [`smtp-http-proxy`](https://github.com/remko/smtp-http-proxy/) and AWS Lambda.

Both Slack and Pushover provide an HTTP API to post notifications to a channel
and mobile devices respectively. To connect HAProxy's SMTP-based alert system to these
APIs, I created a small generic SMTP-to-HTTP bridge `smtp-http-proxy`. This process listens
on a port for incoming SMTP messages, and calls a given HTTP URL in the following 
JSON format:

```json
{
  "envelope": {
    "from": "<sender@example.com>",
    "to": [
      "<receiver1@example.com>",
      "<receiver2@example.com>"
    ]
  },
  "data": "From: sender@example.com\nDate: Sun, 12 Jun 2016 18:03:51 +0200\nSubject: Message\n\nThis is a message"
}
```

Since this is a custom JSON format, it can't be connected directly to the Slack
or Pushover API, but needs to be transformed first. An easy way to do this
transformation without setting up a different server is to create an AWS Lambda
function that accepts incoming HTTP `POST` requests from `smtp-http-proxy`, and
forwards it to the service.

For example, for Slack, you can create an AWS Lambda function (e.g.
`smtp2slack`), implemented in Node.js 4.3, triggered by a `POST` to a new
API Gateway resource (e.g.  `/smtp2slack`). This function gets the input
data as a parameter, and forwards it via HTTP to Slack:

```javascript
var https = require('https');

exports.handler = function (event, context, callback) {
  // Send the entire raw message to Slack.
  // Real code would use an RFC2822 parser such as 
  // https://github.com/andris9/mailparser
  var message = event.data;
  
  var options = {
    hostname: "hooks.slack.com",
    port: 443,
    path: "/services/MY/SLACK/WEBHOOK",
    method: "POST",
    headers: {
      'Content-Type': 'application/json'
    }
  };
  var req = https.request(options, function (res) {
    res.setEncoding('utf8');
    res.on('data', function () {});
    res.on('end', function () {
      callback(res.statusCode === 200 ? null : "Error");
    });
  });
  req.on('error', callback);
  req.write(JSON.stringify({
    username: "HAProxy",
    text: message,
    channel: "#haproxy-errors"
  }));
  req.end();
};
```


Now, we need to run `smtp-http-proxy` with the URL of this API Gateway resource
as target, and in case you configured the API Gateway to require api keys, send
the correct API key with it:

```
./smtp-http-proxy --debug --port 8025 \
  -H "x-api-key: <MY_API_KEY>" \
  --url https://uu71rcz28i.execute-api.eu-central-1.amazonaws.com/prod/smtp2slack
```

Finally, we tell HAProxy that it needs to send its alerts to the `smtp-http-proxy` SMTP 
service listening on port 8025, by editing `haproxy.cfg`:

```
mailers alert-mailers
  mailer smtp1 127.0.0.1:8025

backend mybackend
  email-alert mailers alert-mailers
  email-alert from haproxy@el-tramo.be
  email-alert to haproxy-errors@el-tramo.be
  server mysrv 1.2.3.4:80 check
```

And that's all there is to it. When something bad happens with a HAProxy backend, it will
notify `smtp-http-proxy` via SMTP, which in turn will call the AWS Lambda API, which in
turn will post to the `#haproxy-errors` Slack channel.

For Pushover, all you need is a different AWS Lambda Function to do the transformation:

```javascript
var https = require('https');

exports.handler = function (event, context, callback) {
  // Send the entire raw message to Pushover.
  // Real code would use an RFC2822 parser such as 
  // https://github.com/andris9/mailparser
  var message = event.data;
  
  var options = {
    hostname: "api.pushover.net",
    port: 443,
    path: "/1/messages.json?token=<MYTOKEN>&user=<MYUSER>" 
      + "&message=" + encodeURIComponent(message)
      + "&title=" + encodeURIComponent("HAProxy Alert"),
    method: "POST"
  };
  var req = https.request(options, function (res) {
    res.setEncoding('utf8');
    res.on('data', function () {});
    res.on('end', function () {
      callback(res.statusCode == 200 ? null : "Error");
    });
  });
  req.on('error', callback);
  req.end();
};
```

You can find the final result in the [`smtp-http-proxy` HAProxy Example](https://github.com/remko/smtp-http-proxy/tree/master/examples/haproxy), which contains the sources of 
a Docker image with HAProxy and `smtp-http-proxy`.

