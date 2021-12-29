import re
import sys

refunc = re.compile("^(def )")
respace = re.compile(" *")
INPUT = sys.argv[1]
indent_stack = []
in_comment = False
with open(INPUT) as f:
    s = ""
    for line in f.readlines():
        parts = line.split("_")
        newline = parts[0]
        for w in parts[1:]:
            if len(w) == 0:
                newline += "_"
            elif w[0] in (",", " ", ";"):
                newline += "_" + w
            else:
                newline += w[0].upper() + w[1:]
        if newline[0] == "#":
            newline = "//" + newline[1:]
        newline = newline.replace("'", '"')
        newline = newline.replace("True", "true")
        newline = newline.replace("False", "false")
        newline = newline.replace(" and ", " && ")
        newline = newline.replace(" or ", " || ")
        newline = newline.replace("# ", "// ")
        newline = newline.replace(" elif ", " else if ")
        newline = newline.replace(" in ", " := range ")
        newline = newline.replace("is not None", " != nil ")
        newline = newline.replace("is None", " == nil ")
        newline = newline.replace(" not ", " ! ")
        newline = refunc.sub("func ", newline)

        indent = len(respace.match(newline).group(0))
        if newline.strip() == '"""' or (
                (newline.strip().startswith('"""') or newline.strip().startswith('r"""')) and not newline.strip().endswith('"""')):
            in_comment = not in_comment

        if (not in_comment) and indent_stack:
            while indent_stack and indent < indent_stack[-1]:
                s += " " * indent_stack[-1] + "}\n"
                indent_stack.pop()

            if indent_stack and indent == indent_stack[-1]:
                newline = " " * indent + "} " + newline[indent:]
                indent_stack.pop()

        if (not in_comment) and newline.endswith(":\n"):
            newline = newline[:-2] + " {\n"
            indent_stack.append(indent)

        s += newline

lines = s.split("\n")
re_comment = re.compile('    (r?)"""(.*)"""')
out = []
i = 0
while i < len(lines):
    l = lines[i]
    if l.startswith("func "):
        m = re_comment.match(lines[i+1])
        if m:
            c = m.group(2)
            out.append("// " + c)
            i += 2
        elif lines[i+1].startswith('    """') or lines[i+1].startswith('    r"""'):
            out.append("// " + lines[i+1][7:])
            j = 2
            while i+j < len(lines) and not ('"""' in lines[i+j]):
                if lines[i+j]:
                    out.append("// " + lines[i+j])
                j += 1
            if i+j < len(lines):
                out.append("// " + lines[i+j].replace('"""', ''))
                j += 1
            i = i+j
        else:
            i += 1
        out.append(l)
    else:
        out.append(l)
        i += 1

with open(INPUT, "w") as f:
    f.write("\n".join(out))
