+++
title = "Gotta go fast: making format-on-save fast with prettierd"
draft = true

[taxonomies]
tags = ["javascript", "neovim", "prettier", "typescript"]
+++

Back in June 2020, when I was migrating my Neovim configuration to Lua and to
the native LSP client available in neovim 0.5.0, my main language at work was
TypeScript and we used [prettier](https://prettier.io) to keep our code
formatted, and I had it configured to format-on-save with
[coc-prettier](https://github.com/neoclide/coc-prettier). One of the first
issues I ran into was performance: saving files became deadly slow, to the
point where I gave up and disabled format-on-save.

The thing is: prettier is known to be a faster code formatter, and I didn't
have the issue before, so what's the  problem here? Is Neovim making prettier
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
twice.

The solution is simple: we need a long-running node.js process! If you read
through the issue about slow startups in prettier, someone suggests using [prettier_d](), but after looking at how large that project was, I was a bit scared.

Doing some more research, I found about
[eslint_d.js](https://github.com/mantoni/eslint_d.js/), which solves a similar
issue for eslint, by introducing a daemon which supports binding on a TCP
socket! And the author of eslint_d.js extracted its core functionality in a
library called [core_d.js](https://github.com/mantoni/core_d.js). So I figured
I could combine that library with prettier and make
[prettierd](https://github.com/fsouza/prettierd), a TCP-enabled daemon for
formatting code using prettier!

## Integrating Neovim with prettierd

The code for prettierd is pretty boring, as it is basically a tiny wrapper
around core_d to invoke the proper prettier functions whenever the server
receives a "request".

While ...

## Bonus: using it on the command line with prettierme
