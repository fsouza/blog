import argparse
import sys
from datetime import datetime
from pathlib import Path


def main(args: list[str]) -> int:
    parser = argparse.ArgumentParser(prog="manage.py")
    subparsers = parser.add_subparsers(dest="sub_parser")
    new_post = subparsers.add_parser(
        name="new-post",
        aliases=["new", "new_post"],
        description="starts a new post with the provided slug",
    )
    new_post_options(new_post)
    publish_post = subparsers.add_parser(
        name="publish-post",
        aliases=["publish", "publish_post"],
        description="publishes the post (removes draft & updates date)",
    )
    publish_post_options(publish_post)
    ns = parser.parse_args(args)
    if ns.sub_parser is None:
        parser.print_help()
        return 2
    ns.func(ns)
    return 0


def new_post_options(parser: argparse.ArgumentParser) -> None:
    parser.add_argument("-s", "--slug", dest="slug", required=True)
    parser.set_defaults(func=new_post)


def publish_post_options(parser: argparse.ArgumentParser) -> None:
    parser.add_argument("-p", "--path", dest="path", required=True)
    parser.add_argument("--keep-date", action="store_true", dest="keep_date")
    parser.set_defaults(func=publish_post)


def new_post(args: argparse.Namespace) -> None:
    path = Path(__file__)
    folder = (
        path.parent / "content" / f"{datetime.today().strftime('%Y-%m-%d')}-{args.slug}"
    )
    if folder.exists():
        raise ValueError(f"folder for post already exists: {folder}")
    folder.mkdir(mode=0o755, parents=True)
    content = f"""\
+++
title = "{args.slug}"
draft = true

[taxonomies]
tags = []
+++
"""
    (folder / "index.md").write_text(content, encoding="utf-8")
    print(f"Done! Started new post at {folder}")


def publish_post(args: argparse.Namespace) -> None:
    path = Path(args.path).absolute()
    if not path.exists():
        raise ValueError(f"post not found: {args.path}")
    if not path.is_relative_to(Path(__file__).parent / "content"):
        raise ValueError(f"invalid path: {args.path}. must be under the content dir")

    if path.is_dir():
        post_dir = path
        post_file = path / "index.md"
    else:
        post_dir = path.parent
        post_file = path

    # TODO: update date & remove draft = true


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
