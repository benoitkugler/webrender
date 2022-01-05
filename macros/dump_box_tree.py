from weasyprint.formatting_structure.build import build_formatting_structure
from weasyprint.formatting_structure.boxes import Box
from weasyprint import CSS, HTML, images
from weasyprint.formatting_structure import boxes
from weasyprint.css import get_all_computed_styles
from weasyprint.css.counters import CounterStyle
from weasyprint.css.targets import TargetCollector
import functools
import json
from typing import *


def tree_to_json(box: boxes.Box) -> str:
    assert box.element_tag == 'html'
    assert isinstance(box, boxes.BlockBox)
    assert len(box.children) == 1

    box = box.children[0]
    assert isinstance(box, boxes.BlockBox)
    assert box.element_tag == 'body'

    output = serialize(box.children)
    return json.dumps(output, indent=2)


TEST_UA_STYLESHEET = CSS(filename="html/tree/tests_ua.css")


class FakeHTML(HTML):
    """Like weasyprint.HTML, but with a lighter UA stylesheet."""

    def _ua_stylesheets(self):
        return [TEST_UA_STYLESHEET]


def serializeType(type_name: str) -> int:
    if type_name == "":
        return 0
    elif type_name == "AtomicInlineLevelBox":
        return 1
    elif type_name == "BlockBox":
        return 2
    elif type_name == "BlockContainerBox":
        return 3
    elif type_name == "BlockLevelBox":
        return 4
    elif type_name == "BlockReplacedBox":
        return 5
    elif type_name == "Box":
        return 6
    elif type_name == "FlexBox":
        return 7
    elif type_name == "FlexContainerBox":
        return 8
    elif type_name == "InlineBlockBox":
        return 9
    elif type_name == "InlineBox":
        return 10
    elif type_name == "InlineFlexBox":
        return 11
    elif type_name == "InlineLevelBox":
        return 12
    elif type_name == "InlineReplacedBox":
        return 13
    elif type_name == "InlineTableBox":
        return 14
    elif type_name == "LineBox":
        return 15
    elif type_name == "MarginBox":
        return 16
    elif type_name == "PageBox":
        return 17
    elif type_name == "ParentBox":
        return 18
    elif type_name == "ReplacedBox":
        return 19
    elif type_name == "TableBox":
        return 20
    elif type_name == "TableCaptionBox":
        return 21
    elif type_name == "TableCellBox":
        return 22
    elif type_name == "TableColumnBox":
        return 23
    elif type_name == "TableColumnGroupBox":
        return 24
    elif type_name == "TableRowBox":
        return 25
    elif type_name == "TableRowGroupBox":
        return 26
    elif type_name == "TextBox":
        return 27
    else:
        raise TypeError("invalid type: " + type_name)


def format_tag(box) -> str:
    element_tag = box.element_tag
    if element_tag.startswith("{http://www.w3.org/2000/svg}"):
        element_tag = element_tag[len("{http://www.w3.org/2000/svg}"):]
    return element_tag


def serialize(box_list: [Box]) -> list:
    """Transform a box list into a structure easier to compare for testing."""

    return [{
        "Tag": format_tag(box),
        "Type": serializeType(type(box).__name__),
        # All concrete boxes are either text, replaced, column or parent.
        "Content":
        {"Text": box.text} if isinstance(box, boxes.TextBox)
            else {"Text": '<replaced>'} if isinstance(box, boxes.ReplacedBox)
            else {"C": serialize(
                getattr(box, 'column_groups', ()) + tuple(box.children))}
    } for box in box_list]


def to_go(boxes: List[Any]) -> str:
    code = "[]serBox{"
    for box in boxes:
        tag = box[0]
        type_ = f"{box[1]}BoxT"
        if isinstance(box[2], str):
            content = 'bc{{Text: `{0}`}}'.format(box[2])
        else:
            children = to_go(box[2])
            content = f"bc{{C: {children}}}"
        code += f"""{{"{tag}", {type_}, {content}}},\n"""
    code += "}"
    return code


def build_tree(document: FakeHTML) -> Any:
    counter_style = CounterStyle()
    style_for = get_all_computed_styles(document, counter_style=counter_style)
    get_image_from_uri = functools.partial(
        images.get_image_from_uri, cache={}, url_fetcher=document.url_fetcher,
        optimize_size=(), context=None)
    target_collector = TargetCollector()
    print("Building tree...")
    tree = build_formatting_structure(document.etree_element, style_for, get_image_from_uri, document.base_url,
                                      target_collector, counter_style)
    return tree


def dump_tree(document: FakeHTML) -> str:
    tree = build_tree(document)
    print("Exporting as JSON...")
    return tree_to_json(tree)


# OUTPUT = "resources_test/Wikipedia-Go-expected.json"
# FILENAME = "resources_test/Wikipedia-Go.html"
# BASE_URL = "https://en.wikipedia.org/wiki/Go_(programming_language)"
# print("Loading HTML...")
# document = FakeHTML(filename=FILENAME, base_url=BASE_URL)
# js = dump_tree(document)
# with open(OUTPUT, "w") as fp:
#     fp.write(js)
# print("Done.")

# URL = "https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/preserveAspectRatio"
# OUTPUT = "resources_test/SVG-expected.json"
# print("Loading HTML...")
# document = FakeHTML(url=URL)
# js = dump_tree(document)
# with open(OUTPUT, "w") as fp:
#     fp.write(js)
# print("Done.")


print(to_go(build_tree(FakeHTML(string="""
    <style>
        @page { size: 300px 30px }
        body { margin: 0; background: #fff }
    </style>
    <p><a href="another url"><span>[some url] </span>some content</p>
    """))))
