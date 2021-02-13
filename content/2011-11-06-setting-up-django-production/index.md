+++
title = "Setting up a Django production environment: compiling and configuring nginx"
aliases = ["/2011/11/setting-up-django-production.html"]

[taxonomies]
tags = ["django", "gunicorn", "nginx"]
+++

Here is another series of posts: now I’m going to write about setting up a
[Django](https://djangoproject.com) production environment using
[nginx](https://nginx.org) and [Green Unicorn](http://gunicorn.org) in a
[virtual environment](http://virtualenv.org). The subject in this first post is
nginx, which is my favorite web server.

This post explains how to install nginx from sources, compiling it (on Linux).
You might want to use `apt`, `zif`, `yum` or `ports`, but I prefer building
from sources. So, to build from sources, make sure you have all development
dependencies (C headers, including the PCRE library headers, nginx rewrite
module uses it). If you want to build nginx with SSL support, keep in mind that
you will need the libssl headers too.

Build nginx from source is a straightforward process: all you need to do is
download it from the official site and build with some simple options. In our
setup, we’re going to install nginx under `/opt/nginx`, and use it with the
nginx system user. So, let’s download and extract the latest stable version
(1.0.9) from nginx website:

```
% curl -O http://nginx.org/download/nginx-1.0.9.tar.gz
% tar -xzf nginx-1.0.9.tar.gz
```

Once you have extracted it, just configure, compile and install:

```
% ./configure --prefix=/opt/nginx --user=nginx --group=nginx
% make
% [sudo] make install
```

As you can see, we provided the `/opt/nginx` to configure, make sure the `/opt`
directory exists. Also, make sure that there is a user and a group called
_nginx_, if they don’t exist, add them:

```
% [sudo] adduser --system --no-create-home --disabled-login --disabled-password --group nginx
```

After that, you can start nginx using the command line below:

```
% [sudo] /opt/nginx/sbin/nginx
```

> Linode provides an [init
> script](https://library.linode.com/assets/634-init-deb.sh) that uses
> [start-stop-daemon](http://man.he.net/man8/start-stop-daemon), you might want
> to use it.

#### nginx configuration

nginx comes with a default `nginx.conf` file, let’s change it to reflect the
following configuration requirements:

- nginx should start workers with the `nginx` user
- nginx should have two worker processes
- the PID should be stored in the `/opt/nginx/log/nginx.pid` file
- nginx must have an access log in `/opt/nginx/logs/access.log`
- the configuration for the Django project we’re going to develop should be
  versioned with the entire code, so it must be included in the `nginx.conf`
  file (assume that the `library` project is in the directory `/opt/projects`).

So here is the `nginx.conf` for the requirements above:

```
user  nginx;
worker_processes  2;

pid logs/nginx.pid;

events {
    worker_connections  1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                     '$status $body_bytes_sent "$http_referer" '
                     '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  logs/access.log  main;

    sendfile           on;
    keepalive_timeout  65;

    include /opt/projects/showcase/nginx.conf;
}
```

Now we just need to write the configuration for our Django project. I’m using
an old sample project written while I was working at
[Giran](http://www.giran.com.br): the name is _lojas giranianas_, a nonsense
portuguese joke with a famous brazilian store. It’s an unfinished showcase of
products, it’s like an e-commerce project, but it can’t sell, so it’s just a
product catalog. The code is available at
[Github](https://github.com/fsouza/fast-track-django). The `nginx.conf` file
for the repository is here:

```
server {
    listen 80;
    server_name localhost;

    charset utf-8;

    location / {
        proxy_set_header    X-Real-IP   $remote_addr;
        proxy_set_header    Host        $http_host;
        proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;

        proxy_pass http://localhost:8000;
    }

    location /static {
        root /opt/projects/showcase/;
        expires 1d;
    }
}
```

The server listens on port `80`, responds for the `localhost` hostname ([read
more about the Host
header](https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html)). The
`location /static` directive says that nginx will serve the static files of the
project. It also includes an `expires` directive for caching control. The
`location /` directive makes a `proxy_pass`, forwarding all requisitions to an
upstream server listening on port 8000, this server is the subject of the next
post of the series: the Green Unicorn (gunicorn) server.

Not only the HTTP request itself is forwarded to the gunicorn server, but also
some headers, that helps to properly deal with the request:

* **X-Real-IP:** forwards the remote address to the upstream server, so it can
  know the real IP of the user. When nginx forwards the request to gunicorn,
  without this header, all gunicorn will know is that there is a request coming
  from localhost (or wherever the nginx server is), the remote address is
  always the IP address of the machine where nginx is running (who actually
  make the request to gunicorn)
* **Host:** the `Host` header is forwarded so gunicorn can treat different
  requests for different hosts. Without this header, it will be impossible to
  Gunicorn to have these constraints
* **X-Forwarded-For:** also known as XFF, this header provide more precise
  information about the real IP who makes the request. Imagine there are 10
  proxies between the user machine and your webserver, the XFF header will all
  these proxies comma separated. In order to not turn a proxy into an
  anonymizer, it’s a good practice to always forward this header.

So that is it, in the next post we are going to install and run gunicorn. In
other posts, we’ll see how to make automated deploys using
[Fabric](http://fabfile.org), and some tricks on caching (using the
`proxy_cache` directive and integrating Django, nginx and
[memcached](https://memcached.org)).

See you in next posts.
