+++
title = "Making format-on-save fast with prettierd"

[taxonomies]
tags = ["javascript", "neovim", "prettier", "typescript"]
+++

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Intro](#intro)
- [Who cares?!](#who-cares)
- ["My Visual Studio Code doesn't take a second to format-on-save, your Vim is trash"](#my-visual-studio-code-doesn-t-take-a-second-to-format-on-save-your-vim-is-trash)
- [Installing and starting prettierd](#installing-and-starting-prettierd)
- [Integrating Neovim with prettierd](#integrating-neovim-with-prettierd)
- [Not just TypeScript and JavaScript](#not-just-typescript-and-javascript)
- [Bonus: using it on the command line with prettierme](#bonus-using-it-on-the-command-line-with-prettierme)
- [Feedback](#feedback)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Intro

Back in June of 2020, when I was migrating my Neovim configuration to Lua and
to the native LSP client available in neovim 0.5.0, my main language at work
was TypeScript and we used [prettier](https://prettier.io) to keep our code
formatted, and I had it configured to format-on-save with
[coc-prettier](https://github.com/neoclide/coc-prettier). One of the first
issues I ran into was performance: saving files became deadly slow, to the
point where I gave up and disabled format-on-save.

The thing is: prettier is known to be a fast code formatter, and I didn't have
the issue before, so what's the  problem here? Is Neovim making prettier
slower? Is coc-prettier doing some magic shit?

Before we start looking into this, let's see how prettier behaves when
formatting a somewhat large TypeScript file:

```
％ wc -l sample.ts
     586 sample.ts
％ time npx prettier -w sample.ts
sample.ts 332ms
        0.85 real         0.91 user         0.11 sys
```

It's interesting that prettier reports that it took 332ms to format the file,
but `time` reports that the whole process took 850ms. Who's lying?

Let's take a look at multiple files:

```
％ wc -l sample*.ts
     330 sample1.ts
     718 sample2.ts
     655 sample3.ts
    2511 sample4.ts
     601 sample5.ts
    4815 total
％ time npx prettier -w sample1.ts
sample1.ts 290ms
        0.93 real         0.88 user         0.11 sys
％ time npx prettier -w sample2.ts
sample2.ts 358ms
        1.02 real         1.04 user         0.12 sys
％ time npx prettier -w sample3.ts
sample3.ts 330ms
        0.97 real         0.95 user         0.13 sys
％ time npx prettier -w sample4.ts
sample4.ts 648ms
        1.27 real         1.48 user         0.13 sys
％ time npx prettier -w sample5.ts
sample5.ts 375ms
        1.00 real         1.00 user         0.12 sys
```

Notice how formatting `sample2.ts` and `sample4.ts` takes more than 1 second in
total! Also interesting is the fact that even though `sample1.ts` is less the
half the size of `sample2.ts`, formatting `sample2.ts` does not take twice as
much time.

## Who cares?!

OK, let's take a step back and reflect: who cares if prettier is slow to format
my files? I could run it on a git hook or something like that and not even
notice.

As I mentioned before, I was running format-on-save in Neovim, with a simple
setup, not very fancy:

```vimscript
autocmd BufWritePre *.ts execute "silent %!npx prettier --stdin-filepath '" . expand('%:p') . "'"
```

[(it was a bit fancier than that, but not by much)](https://github.com/fsouza/dotfiles/blob/0b6d3daaa844796f916b3f056a66af0e25a76c3c/autoload/fsouza/prettier.vim#L19-L30)

So, imagine you're using Neovim and every time you save the file you have to
wait 1 second. You'd be mad, right?! There must be a better way...

## "My Visual Studio Code doesn't take a second to format-on-save, your Vim is trash"

To be fair, coc-prettier was pretty fast too. How is that even possible?

Let's go back to our sample files, but this time let's see what prettier does
if we pass all 5 files to it instead of invoking it 5 times:

```
％ time npx prettier -w *.ts
sample1.ts 248ms
sample2.ts 205ms
sample3.ts 110ms
sample4.ts 327ms
sample5.ts 79ms
        1.61 real         2.09 user         0.15 sys
```

This time `sample2.ts` is faster than `sample1.ts`, even though it's twice as
large! What's going on? Turns out [prettier is slow to
start](https://github.com/prettier/prettier/issues/3386), both because of
overhead introduced by node.js and prettier itself (it has tons of plugins and
dependencies).

**And how is it fast to format-on-save using VSCode/coc-prettier?** Simple:
both coc-prettier and Visual Studio Code are long-running node.js processes,
which host prettier as a library, therefore paying the initialization cost
once.

The solution is simple: we need a long-running node.js process! If you read
through the issue about slow startups in prettier, someone suggests using
[prettier_d](https://github.com/josephfrazier/prettier_d), but after looking at
how large that project was, I was a bit scared.

Doing some more research, I found
[eslint_d.js](https://github.com/mantoni/eslint_d.js/), which solves a similar
issue for eslint, by introducing a daemon which supports binding on a TCP
socket! And the author of eslint_d.js extracted its core functionality in a
library called [core_d.js](https://github.com/mantoni/core_d.js). So I figured
I could combine that library with prettier and make
[prettierd](https://github.com/fsouza/prettierd), a TCP-enabled daemon for
formatting code using prettier!

## Installing and starting prettierd

The code for prettierd is pretty boring, as it is basically a tiny wrapper
around core_d to invoke the proper prettier functions whenever the server
receives a "request". The two important things to know about are:

1. You can install it with npm and start it with `prettierd start`:

```
％ npm install -g @fsouza/prettierd
％ prettierd start
```

Alternatively you can do both things with `npx`:

```
％ npx -p @fsouza/prettierd prettierd start
```

2. When it starts, prettierd writes a file with its port number and token

```
％ cat ~/.prettierd
53561 cb2ad753df0aca85
```

This means that prettierd is running on port 53561 and we can use the token
`cb2ad753df0aca85` in our requests to format our source code.

core_d's protocol is pretty simple:


```
<token> <working-dir> <file-name>\n
<file-content>
```

For example, we can use netcat:

```
％ echo "cb2ad753df0aca85 $PWD sample2.ts" | cat - sample2.ts | /usr/bin/time nc localhost 53561 >sample2-formatted.ts
        0.14 real         0.00 user         0.00 sys
```

Remember how formatting `sample2.ts` took over 1 second? Not anymore. :)

## Integrating Neovim with prettierd

Using netcat is great and we could probably write a shell script that we could
use in our (fun fact: someone else did this, check the bonus section!), but
Neovim is powerful enough to connect directly to the TCP server.

How? Neovim has an event loop, which is implemented using
[libuv](https://libuv.org). libuv is probably the best event loop there in the
wild, but don't quote me :) Besides shipping the event loop and all the libuv
code, Neovim also bundles [luv](https://github.com/luvit/luv) and expose the
loop as a Lua API, so we can use `vim.loop.<nice-async-things>`! Taylor
Thompson has written an amazing post about the [using libuv in
Neovim](https://teukka.tech/vimloop.html), go check it out if you're curious :)

Among the utilities provided by libuv, there's a tcp module which includes both
a TCP client and a TCP server! In our case we want to use a client, so we
invoke the `tcp_connect` function:

```lua
local callback = ...
local port = 53561
local token = 'cb2ad753df0aca85'
local client = vim.loop.new_tcp()
vim.loop.tcp_connect(client, '127.0.0.1', port, callback)
```

Since this is async world, we need to pass a callback that gets executed
whenever the connection happens (or in case something goes wrong). So the first
thing we do in our callback is check for errors, which looks familiar for Go
developers:

```lua
local callback = function(err)
  if err then
    error(err)
  end

  ...
end
```

If there are no errors, it means we can send the contents of our file to the
remote server. We have the port and the token, but we also need the contents of
the buffer. So let's grab the contents of the current buffer and send that to
the server, then read back the response and write it back to the buffer! This
time I'll include the entire implementation of the callback, with some inline
comments:

```lua
local callback = function(err)
  if err then
    error(err)
  end

  -- grab the contents of the buffer and add first row to match core_d's protocol
  local bufnr = vim.api.nvim_get_current_buf()
  local first_line = string.format('%s %s %s', token, vim.loop.cwd(), 'sample2.js')
  local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, true)
  table.insert(lines, 1, first_line)

  -- start reading the response
  local response = ''
  vim.loop.read_start(client, function(read_err, chunk)
    -- check if there was any error reading data back, if so, close the
    -- connection and report the error.
    if read_err then
      vim.loop.close(client)
      error('failed to read data from prettierd: ' .. read_err)
    end

    -- libuv will call this callback with no data and no error when it's done,
    -- so if there's data, concatenate it into the final response. Otherwise it
    -- means we're done, so invoke the `write_to_buf` to write the data back.
    if chunk then
      response = response .. chunk
    else
      vim.loop.close(client)
      write_to_buf(response, bufnr)
    end
  end)

  -- write the request
  vim.loop.write(client, table.concat(lines, '\n'))

  -- signal to the server that we're done writing the request
  vim.loop.shutdown(client)
end
```

And here's a simple implementation of `write_to_buf`. The trickiest bit is
error handling: the way errors are reported isn't great, but it's acceptable:
if prettier fails, the last line contains a message in the format `# exit
<code> ...`.

```lua
local function write_to_buf(data, bufnr)
  local new_lines = vim.split(data, '\n')

  -- check for errors
  if string.find(new_lines[#new_lines], '^# exit %d+') then
    error(string.format('failed to format with prettier: %s', data))
  end

  -- write contents
  vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, new_lines)
end
```

Now you can throw all of that in a `format()` function and invoke it on write!

> **Note:** the code here is a simplified version of what I actually use. For
> the actual config, including automatic process management, retries, error
> handling and cursor positioning, checkout
> [prettierd.lua](https://github.com/fsouza/dotfiles/blob/20aa0be6d06224224a50d24c5b63929f16cdb7da/nvim/lua/fsouza/plugin/prettierd.lua#L56)
> in my dotfiles repo.

## Not just TypeScript and JavaScript

Users of prettier are aware of this, but prettier is not just about JavaScript
and TypeScript, it can be used with many other file formats, including HTML,
Markdown, CSS, YAML, JSON and others. Check the [parser configuration in
prettier docs](https://prettier.io/docs/en/options.html#parser) for a full
list, and keep in mind that additional file types can be added via plugins!

## Bonus: using it on the command line with prettierme

If you want to use Vim instead of Neovim, or don't want to maintain a TCP
client in your editor configuration, you can leverage [Ruy Adorno's
prettierme](https://github.com/ruyadorno/prettierme) to use a command line
interface that is more similar to the standard prettier interface. prettierme
is basically a wrapper around our `netcat` example.

## Feedback

Do you have any feedback? Questions? Concerns? Wanna fix a typo? Checkout the
[source for this post in
GitHub](https://github.com/fsouza/blog/blob/HEAD/content/2021-02-12-prettierd-neovim-format-on-save/index.md)
(feel free to send a PR), or the [discussion in the GitHub
repo](https://github.com/fsouza/blog/discussions/14).
