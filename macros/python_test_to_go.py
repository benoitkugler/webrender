import sys
import re

INPUT = sys.argv[1]


ATTRS = [
    "position",
    "width",
    "height",
    "margin",
    "children",
    "border",
    "columnGroups",
    "content",
    "colspan"
]


def export_attibutes(line: str) -> str:
    for attr in ATTRS:
        line = line.replace("." + attr, ".Box()." + attr[0].upper() + attr[1:])
    return line


def unpack_children(line: str) -> str:
    if not ("=" in line and line.endswith(".children\n")):
        return line

    children, box = line.split("=")
    childrenL = [s.strip() for s in children.split(",") if s.strip()]
    box = box[:-10]
    if len(childrenL) == 1:
        return f"{childrenL[0]} := {box}.children[0]\n"
    else:
        return f"{children} := unpack{len(childrenL)}({box})\n"


def replace_triple_quotes(line: str) -> str:
    return line.replace('"""', '`')


def replace_render_pages(line: str) -> str:
    line = line.replace("page, = renderPages(", "page := renderOnePage(t,")
    line = line.replace("pages = renderPages(", "pages := renderPages(t,")
    return line


def replace_assert(line: str) -> str:
    if not ("assert" in line):
        return line

    line = line.strip()
    line = line.replace("assert ", "tu.AssertEqual(t, ")
    line = line.replace("==", ",")

    comment = ""
    if "//" in line:
        i = line.index("//")
        comment = line[i:]
        line = line[:i]

    line = re.sub(r" ([0-9]+)(\.([0-9]+))?",
                  lambda x: "pr.Float(" + x.group(0) + ")", line)

    line = line + ') ' + comment
    return line + "\n"


def replace_xfail(line: str) -> str:
    if line.startswith("@pytest.mark.xfail"):
        return "// xfail"
    return line


def correct_bracket_comment(line: str) -> str:
    return line.replace("} //", "//")


lines = []
with open(INPUT) as f:
    for line in f.readlines():
        line = replace_xfail(line)
        line = correct_bracket_comment(line)
        line = replace_triple_quotes(line)
        line = replace_render_pages(line)
        line = unpack_children(line)
        line = export_attibutes(line)
        line = replace_assert(line)
        lines.append(line)


final_lines = []
for (i, line) in enumerate(lines):
    if not line.startswith("func Test"):
        final_lines.append(line)
        continue

    j = i-1
    while j > 0:
        previous_line = lines[j].strip()
        if previous_line == "":
            j -= 1
            continue
        elif previous_line != "}":
            final_lines.append("}\n")
            break
        else:
            break
    final_lines.append(line)

print("".join(final_lines))
