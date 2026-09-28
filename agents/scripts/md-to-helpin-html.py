#!/usr/bin/env python3
"""Convert pipeline markdown to the HTML Helpin epics need.

create_epic / update_epic store `description` verbatim: markdown shows up as
raw `###` and `**` text. Tasks and Docs convert markdown themselves, so this
is only for epic descriptions.

Covers what epic bodies use: #-headings, paragraphs, -/1. lists, **bold**,
`code`. No external dependencies.

Usage: agents/scripts/md-to-helpin-html.py <file.md>   (or markdown on stdin)
"""
import html
import re
import sys


def inline(text):
    text = html.escape(text, quote=False)
    text = re.sub(r"\*\*(.+?)\*\*", r"<strong>\1</strong>", text)
    return re.sub(r"`(.+?)`", r"<code>\1</code>", text)


def row_cells(line):
    """Split a pipe-table row into its cells, ignoring the outer pipes."""
    return [c.strip() for c in line.strip().strip("|").split("|")]


def is_divider(line):
    """A markdown table's |---|---| separator row."""
    cells = row_cells(line)
    return bool(cells) and all(re.fullmatch(r":?-{2,}:?", c) for c in cells)


def convert(md):
    out, para, lst = [], [], None
    table = None
    fence = None   # language of the open ``` block, or None
    code = []

    def close_para():
        if para:
            out.append("<p>" + inline(" ".join(para)) + "</p>")
            para.clear()

    def close_list():
        nonlocal lst
        if lst:
            out.append(f"</{lst}>")
            lst = None

    def close_table():
        """Emit the buffered rows as a real table, first row as the header."""
        nonlocal table
        if not table:
            table = None
            return
        head, body = table[0], table[1:]
        out.append("<table><tbody>")
        out.append("<tr>" + "".join(
            f"<th><p>{inline(c)}</p></th>" for c in head) + "</tr>")
        for r in body:
            r = (r + [""] * len(head))[:len(head)]
            out.append("<tr>" + "".join(
                f"<td><p>{inline(c)}</p></td>" for c in r) + "</tr>")
        out.append("</tbody></table>")
        table = None

    for line in md.strip().splitlines():
        stripped = line.strip()
        if stripped.startswith("```"):
            if fence is None:
                close_para()
                close_list()
                close_table()
                fence = stripped[3:].strip()
                code = []
            else:
                lang = f' class="language-{fence}"' if fence else ""
                body = html.escape("\n".join(code), quote=False)
                out.append(f"<pre><code{lang}>{body}</code></pre>")
                fence = None
            continue
        if fence is not None:
            code.append(line)
            continue
        if not line.strip() or line.strip() == "---":
            close_para()
            close_list()
            close_table()
            continue
        heading = re.match(r"^(#{1,6}) (.*)", line)
        if heading:
            close_para()
            close_list()
            close_table()
            n = len(heading.group(1))
            out.append(f"<h{n}>{inline(heading.group(2))}</h{n}>")
            continue
        if line.lstrip().startswith("|"):
            close_para()
            close_list()
            if is_divider(line):
                continue
            if table is None:
                table = []
            table.append(row_cells(line))
            continue
        close_table()
        bullet = re.match(r"^- (?:\[[ x]\] )?(.*)", line)
        number = re.match(r"^\d+\. (.*)", line)
        if bullet or number:
            close_para()
            kind = "ul" if bullet else "ol"
            if lst != kind:
                close_list()
                out.append(f"<{kind}>")
                lst = kind
            out.append("<li><p>" + inline((bullet or number).group(1)) + "</p></li>")
            continue
        close_list()
        para.append(line.strip())
    close_para()
    close_list()
    close_table()
    return "\n".join(out)


if __name__ == "__main__":
    src = open(sys.argv[1]).read() if len(sys.argv) > 1 else sys.stdin.read()
    print(convert(src))
