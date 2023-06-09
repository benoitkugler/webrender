import source_box
import typing
import inspect

# the following interfaces need additional methods
# which must be completed manually
CLASS_NEED_METHODS = [
    "BlockLevelBox",
    "ReplacedBox",
    "TableBox",
]

# the followings types have not concrete struct associated
ABSTRACT_TYPES = [
    "Box",
    "BlockContainerBox",
    "FlexContainerBox",
    "ParentBox",
    "AtomicInlineLevelBox",
    "BlockLevelBox",
    "InlineLevelBox",
]


def get_class_comment(c: type) -> str:
    python_comment = inspect.getdoc(c) or ""
    out = python_comment.replace("\n\n", "\n").replace('"""', "")
    return "\n".join("// " + line for line in out.split("\n"))


def get_parent_classes(class_: type) -> typing.List[str]:
    return sorted(set(c.__name__ for c in class_.__bases__ if c.__name__ != "object"))


# the root class Box is not returned
def resolve_ancestors(class_: type) -> typing.Set[str]:
    level = [c for c in class_.__bases__ if c.__name__ !=
             "object" and c.__name__ != "Box"]
    out: typing.Set[str] = set(c.__name__ for c in level)
    for parent in level:
        out = out.union(resolve_ancestors(parent))
    return out


# Returns true if a standard AnonymousFrom method may be generated for this type
def has_default_anonymous_from_method(class_: type) -> bool:
    if repr(inspect.signature(getattr(class_, "__init__"))) != "<Signature (self, element_tag, style, element, children)>":
        return False
    class_name = class_.__name__
    for name, func in inspect.getmembers(class_, inspect.ismethod):
        if name == "anonymous_from" and func.__qualname__.split(".")[0] == class_name:
            return False  # own_anonymous_from = True

    return True


def should_generate_anonymous_from(class_: type) -> bool:
    return (class_.__name__ not in ABSTRACT_TYPES) and has_default_anonymous_from_method(class_)


""" Generates the interface for the given box class """


def get_itf_and_type_code(class_: type) -> str:
    comment = get_class_comment(c)

    class_name = c.__name__

    # special case for the abstract class which is handled manually
    if class_name == "Box":
        return ""

    type_methods = [f'{parent}ITF' for parent in get_parent_classes(
        c)] + [f"is{class_name} ()"]

    if class_name in CLASS_NEED_METHODS:
        type_methods.append(f"methods{class_name}")

    s = "\n".join(type_methods)

    itf_code = f"""
        {comment}
        type {class_name}ITF interface {{
            {s}
        }}

        """

    if not class_name in ABSTRACT_TYPES:
        itf_code += f"""func ({class_name}) Type() BoxType {{ return {class_name[:-3]}T }}
        """

        itf_code += f"""func (b *{class_name}) Box() *BoxFields {{ return &b.BoxFields }}
        """

        itf_code += f"""func (b {class_name}) Copy() Box {{ return &b }}
        """

        itf_code += f"""func ({class_name}) IsClassicalBox() bool {{ return true }}
        """

        itf_code += f"""func ({class_name}) is{class_name}() {{ }}
        """

        for ancestor in sorted(resolve_ancestors(class_)):
            itf_code += f"""func({class_name}) is{ancestor}() {{}}
            """

    if should_generate_anonymous_from(class_):
        itf_code += f"""
        func {class_name}AnonymousFrom(parent Box, children []Box) *{class_name} {{
            style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil)
            out := New{class_name}(style, parent.Box().Element, parent.Box().PseudoType, children)
            return out
        }}
        """

    return itf_code


CLASSES: typing.List[type] = [obj for name, obj in inspect.getmembers(
    source_box) if inspect.isclass(obj)]


output = """package boxes

// Code generated by macros/boxes.py DO NOT EDIT

import "github.com/benoitkugler/webrender/html/tree"

"""

type_constants = """
// BoxType represents a box type.
type BoxType uint8

const(
    invalidType BoxType=iota
    {types}
)
"""

type_is_instance = """
// Returns true is the box is an instance of t.
func(t BoxType) IsInstance(box BoxITF) bool {{
    var isInstance bool
    switch t {{
        {switches}}}
    return isInstance
}}
"""

type_string = """
func(t BoxType) String() string {{
    switch t {{
        {type_strings}}}
    return "<invalid box type>"
}}
"""

compilation_checks = """
var(
    {checks}
)
"""

type_anonymous = """
func(t BoxType) AnonymousFrom(parent Box, children []Box) Box {{
    switch t {{
        {type_anonymous_switches}}}
    return nil
}}
"""

types, switches, checks, type_strings, type_anonymous_switches = "", "", "", "", ""
for c in CLASSES:
    output += get_itf_and_type_code(c)
    class_name = c.__name__
    type_name = class_name[:-3] + "T"
    # Generate the type value for the given class
    types += f"""{type_name}\n"""
    switches += f"""case {type_name}:
        _, isInstance = box.({class_name}ITF)
    """
    type_strings += f"""case {type_name}:
        return "{class_name}"
    """

    if not class_name in ABSTRACT_TYPES:
        checks += f"_ {class_name}ITF = (*{class_name})(nil)\n"

    if should_generate_anonymous_from(c):
        type_anonymous_switches += f"""case {type_name}:
        return {class_name}AnonymousFrom(parent, children)
    """

output += type_constants.format(types=types)
output += type_is_instance.format(switches=switches)
output += type_string.format(type_strings=type_strings)
output += compilation_checks.format(checks=checks)
output += type_anonymous.format(
    type_anonymous_switches=type_anonymous_switches)

print(output)
