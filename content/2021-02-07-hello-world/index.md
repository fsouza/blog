+++
title = "Hello World!"

[taxonomies]
tags = ["blogging", "testing"]
+++

This is my first post.

And I am trying to test Zola, to confirm it's a good match! Who knows?!

Here's some Python to test the code block:

```python
import os
import sys


def main() -> int:
    home = os.getenv("HOME")
    if home is None:
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())
```
