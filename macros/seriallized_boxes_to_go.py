# This script is used to converted Python box tree to test references
import pyperclip
from typing import *


def to_go(boxes: List[Any]) -> str:
    code = "[]SerBox{"
    for box in boxes:
        tag = box[0]
        type_ = f"{box[1]}BoxT"
        if isinstance(box[2], str):
            content = 'BC{{Text: `{0}`}}'.format(box[2])
        else:
            children = to_go(box[2])
            content = f"BC{{C: {children}}}"
        code += f"""{{"{tag}", {type_}, {content}}},\n"""
    code += "}"
    return code


IN = []

with open("/home/benoit/Téléchargements/WeasyPrint/tmp") as f:
    l = "tmp = " + f.read()
    loc: Dict[str, Any] = {}
    exec(l, globals(), loc)
    IN = loc["tmp"]

pyperclip.copy(to_go(IN))
print("Copied in clipboard.")
